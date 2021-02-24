package middleware

import (
	"fmt"
	customio "github.com/Layer9Berlin/pipedream/src/custom/io"
	"github.com/Layer9Berlin/pipedream/src/custom/math"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/parsing"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/hashicorp/go-multierror"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ExecutionContext is the data model keeping track of everything required to execute a pipeline file
type ExecutionContext struct {
	// SelectableFiles is a list of filenames of all pipeline files in the current directory
	//
	// Only one file can be selected for execution, but all files will be parsed.
	SelectableFiles []string
	// PipelineFiles is a list of pipeline files in the current directory, including recursive imports
	PipelineFiles []pipeline.File
	// Definitions contains all pipeline definitions in the PipelineFiles, as well as built-in pipes
	Definitions pipeline.DefinitionsLookup
	// MiddlewareStack is a list of middleware items that will be executed in turn
	MiddlewareStack []Middleware
	// Defaults contains some execution options that can be set at file level
	Defaults pipeline.DefaultSettings
	// Hooks can execute certain functions before or after pipeline execution
	//
	// Currently not implemented or used.
	Hooks pipeline.HookDefinitions

	// Log is the execution context's logger
	Log *logrus.Logger

	// ProjectPath is the directory in which the tool is currently executing
	ProjectPath string
	// RootFileName is the name of the file selected for execution
	RootFileName string

	rootRun *pipeline.Run

	Runs        []*pipeline.Run
	Connections []*pipeline.DataConnection

	errors *multierror.Error

	interruptChannel chan os.Signal

	preCallback       func(*pipeline.Run)
	postCallback      func(*pipeline.Run)
	executionFunction func(*pipeline.Run)

	parser *parsing.Parser

	// UserPromptImplementation by default shows an interactive prompt to the user
	//
	// Can be overwritten if you need a different implementation e.g. for tests.
	UserPromptImplementation func(
		label string,
		items []string,
		initialSelection int,
		size int,
		input io.ReadCloser,
		output io.WriteCloser,
	) (int, string, error)
}

// NewExecutionContext creates a new ExecutionContext with the specified options
func NewExecutionContext(options ...ExecutionContextOption) *ExecutionContext {
	executionContext := &ExecutionContext{
		parser:                   parsing.NewParser(),
		Log:                      logrus.New(),
		UserPromptImplementation: defaultUserPrompt,
	}
	executionContext.executionFunction = func(run *pipeline.Run) {
		executionContext.unwindStack(run, 0)
	}
	for _, option := range options {
		option(executionContext)
	}
	return executionContext
}

// CancelAll cancels all active runs
func (executionContext *ExecutionContext) CancelAll() error {
	err := &multierror.Error{}
	for _, run := range executionContext.Runs {
		err = multierror.Append(err, run.Cancel())
	}
	return err.ErrorOrNil()
}

// FullRun starts the complete execution procedure for a nested pipeline, unwinding the entire middleware stack again
func (executionContext *ExecutionContext) FullRun(options ...FullRunOption) *pipeline.Run {
	runOptions := FullRunOptions{}
	for _, option := range options {
		option(&runOptions)
	}
	var pipelineDefinition *pipeline.Definition = nil
	if runOptions.pipelineIdentifier != nil {
		if definition, ok := LookUpPipelineDefinition(executionContext.Definitions, *runOptions.pipelineIdentifier, executionContext.RootFileName); ok {
			pipelineDefinition = definition
		}
	}
	pipelineRun, err := pipeline.NewRun(runOptions.pipelineIdentifier, runOptions.arguments, pipelineDefinition, runOptions.parentRun)
	if err != nil {
		panic(fmt.Errorf("failed to create pipeline run: %w", err))
	}
	pipelineRun.Log.ErrorCallback = func(err error) {
		executionContext.errors = multierror.Append(executionContext.errors, err)
	}
	if runOptions.logWriter == nil {
		if runOptions.parentRun != nil {
			runOptions.parentRun.Log.Debug(
				fields.Symbol("🏃"),
				fields.Message("full run"),
				fields.Info(pipelineRun.String()),
				fields.Color("cyan"),
			)
			runOptions.parentRun.Log.AddReaderEntry(pipelineRun.Log)
		}
	} else {
		if pipelineRun.Log.Level() >= logrus.DebugLevel {
			indentation := math.MaxInt(pipelineRun.Log.Indentation-2, 0)
			initialLogData, _ := logging.LogFormatter{}.Format(
				fields.EntryWithFields(
					fields.Symbol("🏃"),
					fields.Message("full run"),
					fields.Info(pipelineRun.String()),
					fields.Color("cyan"),
					fields.Indentation(indentation),
				))
			go func() {
				_, _ = runOptions.logWriter.Write(initialLogData)
				_, _ = io.Copy(runOptions.logWriter, pipelineRun.Log)
				_ = runOptions.logWriter.Close()
			}()
		} else {
			go func() {
				_, _ = io.Copy(runOptions.logWriter, pipelineRun.Log)
				_ = runOptions.logWriter.Close()
			}()
		}
	}
	pipelineRun.Log.Debug(
		fields.Message("starting"),
		fields.Info(pipelineRun.String()),
		fields.Symbol("▶️"),
		fields.Color("green"),
	)
	if runOptions.parentRun == nil && executionContext.rootRun == nil {
		executionContext.rootRun = pipelineRun
	}
	executionContext.Runs = append(executionContext.Runs, pipelineRun)
	if runOptions.preCallback != nil {
		runOptions.preCallback(pipelineRun)
	}
	executionContext.executionFunction(pipelineRun)
	if runOptions.postCallback != nil {
		runOptions.postCallback(pipelineRun)
	}
	// copy the stderr output after all other middleware has processed it
	stderrCopy := pipelineRun.Stderr.Copy()
	// don't close the run's log until we are done writing to it
	pipelineRun.LogClosingWaitGroup.Add(1)
	go func() {
		// read the entire remaining stderr
		stderr, _ := ioutil.ReadAll(stderrCopy)
		// if there is any output, log it!
		if len(stderr) > 0 {
			pipelineRun.Log.StderrOutput(string(stderr))
		}
		// now the run can complete
		pipelineRun.LogClosingWaitGroup.Done()
	}()
	pipelineRun.LogClosingWaitGroup.Add(1)
	go func() {
		pipelineRun.Stdin.Wait()
		stdin := pipelineRun.Stdin.String()
		if len(stdin) > 0 {
			pipelineRun.Log.Trace(
				fields.Message("input"),
				fields.Info(stdin),
				fields.Symbol("↘️️"),
				fields.Color("gray"),
			)
		}
		pipelineRun.LogClosingWaitGroup.Done()
	}()
	pipelineRun.LogClosingWaitGroup.Add(1)
	go func() {
		pipelineRun.Stdout.Wait()
		stdout := pipelineRun.Stdout.String()
		if len(stdout) > 0 {
			pipelineRun.Log.Trace(
				fields.Message("output"),
				fields.Info(stdout),
				fields.Symbol("↗️"),
				fields.Color("gray"),
			)
		}
		pipelineRun.LogClosingWaitGroup.Done()
	}()
	pipelineRun.Close()
	return pipelineRun
}

// PipelineFileAtPath returns a *pipeline.File corresponding to the parsed pipeline file at the given path, if any
func (executionContext *ExecutionContext) PipelineFileAtPath(path string) (*pipeline.File, error) {
	for _, file := range executionContext.PipelineFiles {
		if file.Path == path {
			return &file, nil
		}
	}
	return nil, fmt.Errorf("file not found")
}

// LookUpPipelineDefinition looks for a particular pipeline within the already parsed deinitions
func LookUpPipelineDefinition(definitionsLookup pipeline.DefinitionsLookup, identifier string, rootFileName string) (*pipeline.Definition, bool) {
	definitions, ok := definitionsLookup[identifier]
	if !ok {
		return nil, false
	}
	// pipelines defined in the same file take precedence,
	// then the first public pipeline
	// private pipelines in other files will only be invoked if there is no public match
	var firstPublicMatch *pipeline.Definition = nil
	var firstPrivateMatch *pipeline.Definition = nil
	for _, definition := range definitions {
		// need to copy here to prevent modification of saved result by for loop
		definitionCopy := definition
		if rootFileName != "" && definition.FileName == rootFileName {
			return &definitionCopy, true
		} else if firstPublicMatch == nil && definition.Public {
			firstPublicMatch = &definitionCopy
		} else if firstPrivateMatch == nil && !definition.Public {
			firstPrivateMatch = &definitionCopy
		}
	}
	if firstPublicMatch != nil {
		return firstPublicMatch, true
	} else if firstPrivateMatch != nil {
		return firstPrivateMatch, true
	}
	return nil, false
}

func (executionContext *ExecutionContext) unwindStack(
	pipelineRun *pipeline.Run,
	currentIndex int,
) {
	if len(executionContext.MiddlewareStack) > currentIndex {
		currentMiddleware := executionContext.MiddlewareStack[currentIndex]
		currentMiddleware.Apply(pipelineRun, func(newRun *pipeline.Run) {
			executionContext.unwindStack(newRun, currentIndex+1)
		}, executionContext)
	}
}

// Execute runs a pipeline and outputs the result
func (executionContext *ExecutionContext) Execute(pipelineIdentifier string, stdoutWriter io.Writer, stderrWriter io.Writer) {

	fullRun := executionContext.FullRun(WithIdentifier(&pipelineIdentifier))
	fullRun.Close()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = io.Copy(stdoutWriter, fullRun.Log)
		waitGroup.Done()
	}()
	fullRun.Wait()

	waitGroup.Wait()
	outputResult(fullRun, stdoutWriter)
	outputErrors(executionContext.errors, stderrWriter)
}

// SetUpPipelines collects and parses all relevant pipeline files
func (executionContext *ExecutionContext) SetUpPipelines(fileFlag string) error {
	executionContext.Log.Tracef("Setting up pipelines...")

	filePaths, err := executionContext.parser.BuiltInPipelineFilePaths(executionContext.ProjectPath)
	if err != nil {
		return err
	}

	_, builtInPipelineDefinitions, _, err := executionContext.parser.ParsePipelineFiles(filePaths, true)
	if err != nil {
		return err
	}

	localPipelineFilePaths, err := executionContext.parser.UserPipelineFilePaths(fileFlag)
	if err != nil {
		return err
	}

	pipelineFilePathsIncludingImported, err := executionContext.parser.RecursivelyAddImports(localPipelineFilePaths)
	if err != nil {
		return err
	}
	executionContext.Log.Tracef("Found pipeline files:\n - %v", strings.Join(pipelineFilePathsIncludingImported, "\n - "))

	defaults, userPipelineDefinitions, files, err := executionContext.parser.ParsePipelineFiles(pipelineFilePathsIncludingImported, false)
	if err != nil {
		return err
	}

	executionContext.SelectableFiles = localPipelineFilePaths
	executionContext.Defaults = defaults
	executionContext.Definitions = pipeline.MergePipelineDefinitions(builtInPipelineDefinitions, userPipelineDefinitions)
	executionContext.PipelineFiles = files

	return nil
}

func defaultUserPrompt(
	label string,
	items []string,
	initialSelection int,
	size int,
	input io.ReadCloser,
	output io.WriteCloser,
) (int, string, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		CursorPos: initialSelection,
		Size:      size,
		Stdin:     input,
		Stdout:    customio.NewBellSkipper(output),
	}
	return prompt.Run()
}

// SetUpCancelHandler registers a handler for interrupt signals
func (executionContext *ExecutionContext) SetUpCancelHandler(handler func(), writer io.Writer) {
	if executionContext.interruptChannel == nil {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-signalChannel
			_, _ = io.WriteString(writer, "\nExecution cancelled...\n")
			handler()
			close(signalChannel)
			signal.Reset(os.Interrupt, syscall.SIGTERM)
		}()
		executionContext.interruptChannel = signalChannel
	}
}

// WaitForRun blocks until a run with the specified identifier is found and completes
func (executionContext *ExecutionContext) WaitForRun(identifier string) *pipeline.Run {
	for {
		for _, run := range executionContext.Runs {
			if run != nil && run.Identifier != nil && *run.Identifier == identifier {
				run.Wait()
				return run
			}
		}
		time.Sleep(250)
	}
}

// UserRuns lists all runs of pipes that are not built-in
func (executionContext *ExecutionContext) UserRuns() []*pipeline.Run {
	result := make([]*pipeline.Run, 0, len(executionContext.Runs))
	for _, run := range executionContext.Runs {
		if run.Definition == nil || !run.Definition.BuiltIn {
			result = append(result, run)
		}
	}
	return result
}
