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
	"syscall"
)

func WithUserPromptImplementation(implementation func(
	label string,
	items []string,
	initialSelection int,
	size int,
	input io.ReadCloser,
	output io.WriteCloser,
) (int, string, error)) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.UserPromptImplementation = implementation
	}
}

type ExecutionContext struct {
	PipelineFiles   []pipeline.File
	Definitions     pipeline.PipelineDefinitionsLookup
	MiddlewareStack []Middleware
	Defaults        pipeline.DefaultSettings
	Hooks           pipeline.HookDefinitions

	Log *logrus.Logger

	ProjectPath  string
	RootFileName string

	rootRun *pipeline.Run

	SelectableFiles []string

	Runs []*pipeline.Run

	errors *multierror.Error

	preCallback       func(*pipeline.Run)
	postCallback      func(*pipeline.Run)
	executionFunction func(*pipeline.Run)

	parser *parsing.Parser

	UserPromptImplementation func(
		label string,
		items []string,
		initialSelection int,
		size int,
		input io.ReadCloser,
		output io.WriteCloser,
	) (int, string, error)
}

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

func (executionContext *ExecutionContext) CancelAll() error {
	err := &multierror.Error{}
	for _, run := range executionContext.Runs {
		err = multierror.Append(err, run.Cancel())
	}
	return err.ErrorOrNil()
}

func (executionContext *ExecutionContext) FullRun(options ...FullRunOption) *pipeline.Run {
	runOptions := FullRunOptions{}
	for _, option := range options {
		option(&runOptions)
	}
	var pipelineDefinition *pipeline.PipelineDefinition = nil
	if runOptions.pipelineIdentifier != nil {
		if definition, ok := LookUpPipelineDefinition(executionContext.Definitions, *runOptions.pipelineIdentifier, executionContext.RootFileName); ok {
			pipelineDefinition = definition
		}
	}
	pipelineRun, err := pipeline.NewPipelineRun(runOptions.pipelineIdentifier, runOptions.arguments, pipelineDefinition, runOptions.parentRun)
	if err != nil {
		panic(fmt.Errorf("failed to create pipeline run: %w", err))
	}
	pipelineRun.Log.ErrorCallback = func(err error) {
		executionContext.errors = multierror.Append(executionContext.errors, err)
	}
	if runOptions.logWriter == nil {
		if runOptions.parentRun != nil {
			runOptions.parentRun.Log.DebugWithFields(
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
			initialLogData, _ := logging.CustomFormatter{}.Format(
				fields.EntryWithFields(
					fields.Symbol("🏃"),
					fields.Message("full run"),
					fields.Info(pipelineRun.String()),
					fields.Color("cyan"),
					fields.Indentation(indentation),
				))
			go func() {
				_, _ = runOptions.logWriter.Write(initialLogData)
				pipelineRun.Wait()
				_, _ = io.Copy(runOptions.logWriter, pipelineRun.Log)
				_ = runOptions.logWriter.Close()
			}()
		} else {
			go func() {
				pipelineRun.Wait()
				_, _ = io.Copy(runOptions.logWriter, pipelineRun.Log)
				_ = runOptions.logWriter.Close()
			}()
		}
	}
	pipelineRun.Log.DebugWithFields(
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
			pipelineRun.Log.TraceWithFields(
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
			pipelineRun.Log.TraceWithFields(
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

func (executionContext *ExecutionContext) PipelineFileAtPath(path string) (*pipeline.File, error) {
	for _, file := range executionContext.PipelineFiles {
		if file.FileName == path {
			return &file, nil
		}
	}
	return nil, fmt.Errorf("file not found")
}

func LookUpPipelineDefinition(definitionsLookup pipeline.PipelineDefinitionsLookup, identifier string, rootFileName string) (*pipeline.PipelineDefinition, bool) {
	definitions, ok := definitionsLookup[identifier]
	if !ok {
		return nil, false
	}
	// pipelines defined in the same file take precedence,
	// then the first public pipeline
	// private pipelines in other files will only be invoked if there is no public match
	var firstPublicMatch *pipeline.PipelineDefinition = nil
	var firstPrivateMatch *pipeline.PipelineDefinition = nil
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

func (executionContext *ExecutionContext) Execute(pipelineIdentifier string, stdoutWriter io.Writer, stderrWriter io.Writer) {

	fullRun := executionContext.FullRun(WithIdentifier(&pipelineIdentifier))
	fullRun.Close()
	setUpCancelHandler(func() {
		err := executionContext.Cancel()
		if err != nil {
			fmt.Printf("Failed to cancel: %v", err)
		}
	})
	go func() {
		_, _ = io.Copy(stdoutWriter, fullRun.Log)
	}()
	fullRun.Wait()

	outputResult(fullRun, stdoutWriter)
	outputErrors(executionContext.errors, stderrWriter)
}

func (executionContext *ExecutionContext) SetUpPipelines(args []string) error {
	executionContext.Log.Tracef("Setting up pipelines...")

	filePaths, err := executionContext.parser.BuiltInPipelineFilePaths(executionContext.ProjectPath)
	if err != nil {
		return err
	}

	_, builtInPipelineDefinitions, _, err := executionContext.parser.ParsePipelineFiles(filePaths, true)
	if err != nil {
		return err
	}

	localPipelineFilePaths, err := executionContext.parser.UserPipelineFilePaths(args)
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

func setUpCancelHandler(handler func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		fmt.Println("\nExecution cancelled...")
		handler()
		close(signalChannel)
		signal.Reset(os.Interrupt, syscall.SIGTERM)
	}()
}

func (executionContext *ExecutionContext) Cancel() error {
	err := &multierror.Error{}
	for _, run := range executionContext.Runs {
		if !run.Completed() {
			err = multierror.Append(err, run.Cancel())
		}
	}
	return err.ErrorOrNil()
}
