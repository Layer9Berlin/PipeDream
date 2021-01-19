package middleware

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"io"
	"pipedream/src/helpers/custom_math"
	"pipedream/src/logging"
	"pipedream/src/logging/log_fields"
	"pipedream/src/models"
	"pipedream/src/parsers"
	"strings"
)

type ExecutionContextOption func(*ExecutionContext)

func WithExecutionFunction(executionFunction func(run *models.PipelineRun)) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.executionFunction = executionFunction
	}
}

func WithDefinitionsLookup(definitions models.PipelineDefinitionsLookup) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.Definitions = definitions
	}
}

func WithProjectPath(projectPath string) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.ProjectPath = projectPath
	}
}

func WithMiddlewareStack(stack []Middleware) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.MiddlewareStack = stack
	}
}

func WithActivityIndicator(activityIndicator *logging.NestedActivityIndicator) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.ActivityIndicator = activityIndicator
	}
}

func WithParser(parser *parsers.Parser) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.parser = parser
	}
}

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
	PipelineFiles   []models.PipelineFile
	Definitions     models.PipelineDefinitionsLookup
	MiddlewareStack []Middleware
	Defaults        models.DefaultSettings
	Hooks           models.HookDefinitions

	Log *logrus.Logger

	ProjectPath  string
	RootFileName string

	rootRun *models.PipelineRun

	SelectableFiles []string

	runs []*models.PipelineRun

	ActivityIndicator *logging.NestedActivityIndicator

	preCallback       func(*models.PipelineRun)
	postCallback      func(*models.PipelineRun)
	executionFunction func(*models.PipelineRun)

	parser *parsers.Parser

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
		parser: parsers.NewParser(),
		Log:    logrus.New(),
		UserPromptImplementation: defaultUserPrompt,
	}
	executionContext.executionFunction = func(run *models.PipelineRun) {
		executionContext.unwindStack(run, 0)
	}
	for _, option := range options {
		option(executionContext)
	}
	return executionContext
}

func (executionContext *ExecutionContext) CancelAll() error {
	err := &multierror.Error{}
	for _, run := range executionContext.runs {
		err = multierror.Append(err, run.Cancel())
	}
	return err.ErrorOrNil()
}

type FullRunOptions struct {
	arguments          map[string]interface{}
	logWriter          io.WriteCloser
	parentRun          *models.PipelineRun
	pipelineIdentifier *string
	postCallback       func(*models.PipelineRun)
	preCallback        func(*models.PipelineRun)
}

type FullRunOption func(*FullRunOptions)

func WithIdentifier(identifier *string) FullRunOption {
	return func(options *FullRunOptions) {
		options.pipelineIdentifier = identifier
	}
}

func WithParentRun(parentRun *models.PipelineRun) FullRunOption {
	return func(options *FullRunOptions) {
		options.parentRun = parentRun
	}
}

func WithLogWriter(logWriter io.WriteCloser) FullRunOption {
	return func(options *FullRunOptions) {
		options.logWriter = logWriter
	}
}

func WithArguments(arguments map[string]interface{}) FullRunOption {
	return func(options *FullRunOptions) {
		options.arguments = arguments
	}
}

func WithSetupFunc(preCallback func(*models.PipelineRun)) FullRunOption {
	return func(options *FullRunOptions) {
		options.preCallback = preCallback
	}
}

func WithTearDownFunc(postCallback func(*models.PipelineRun)) FullRunOption {
	return func(options *FullRunOptions) {
		options.postCallback = postCallback
	}
}

func (executionContext *ExecutionContext) FullRun(options ...FullRunOption) *models.PipelineRun {
	runOptions := FullRunOptions{}
	for _, option := range options {
		option(&runOptions)
	}
	var pipelineDefinition *models.PipelineDefinition = nil
	if runOptions.pipelineIdentifier != nil {
		if definition, ok := LookUpPipelineDefinition(executionContext.Definitions, *runOptions.pipelineIdentifier, executionContext.RootFileName); ok {
			pipelineDefinition = definition
		}
	}
	run, err := models.NewPipelineRun(runOptions.pipelineIdentifier, runOptions.arguments, pipelineDefinition, runOptions.parentRun)
	if err != nil {
		panic(fmt.Errorf("failed to create pipeline run: %w", err))
	}
	if runOptions.logWriter == nil {
		if runOptions.parentRun != nil {
			runOptions.parentRun.Log.DebugWithFields(
				log_fields.Symbol("🏃"),
				log_fields.Message("full run"),
				log_fields.Info(run.String()),
				log_fields.Color("cyan"),
			)
			runOptions.parentRun.Log.AddReaderEntry(run.Log.Reader())
		}
	} else if run.Log.Level() >= logrus.DebugLevel {
		indentation := custom_math.MaxInt(run.Log.Indentation-2, 0)
		initialLogData, _ := logging.CustomFormatter{}.Format(
			log_fields.EntryWithFields(
				log_fields.Symbol("🏃"),
				log_fields.Message("full run"),
				log_fields.Info(run.String()),
				log_fields.Color("cyan"),
				log_fields.Indentation(indentation),
			))
		go func() {
			_, _ = runOptions.logWriter.Write(initialLogData)
			run.Wait()
			_, _ = runOptions.logWriter.Write(run.Log.Bytes())
			_ = runOptions.logWriter.Close()
		}()
	}
	run.Log.DebugWithFields(
		log_fields.Message("starting"),
		log_fields.Info(run.String()),
		log_fields.Symbol("▶️"),
		log_fields.Color("green"),
	)
	if runOptions.parentRun == nil && executionContext.rootRun == nil {
		executionContext.rootRun = run
	}
	executionContext.runs = append(executionContext.runs, run)
	if executionContext.ActivityIndicator != nil {
		executionContext.ActivityIndicator.AddSpinner(run, run.Log.Indentation)
	}
	if runOptions.preCallback != nil {
		runOptions.preCallback(run)
	}
	executionContext.executionFunction(run)
	if runOptions.postCallback != nil {
		runOptions.postCallback(run)
	}
	run.Close()
	return run
}

func (executionContext *ExecutionContext) PipelineFileAtPath(path string) (*models.PipelineFile, error) {
	for _, file := range executionContext.PipelineFiles {
		if file.FileName == path {
			return &file, nil
		}
	}
	return nil, fmt.Errorf("file not found")
}

func LookUpPipelineDefinition(definitionsLookup models.PipelineDefinitionsLookup, identifier string, rootFileName string) (*models.PipelineDefinition, bool) {
	definitions, ok := definitionsLookup[identifier]
	if !ok {
		return nil, false
	}
	// pipelines defined in the same file take precedence,
	// then the first public pipeline
	// private pipelines in other files will only be invoked if there is no public match
	var firstPublicMatch *models.PipelineDefinition = nil
	var firstPrivateMatch *models.PipelineDefinition = nil
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
	run *models.PipelineRun,
	currentIndex int,
) {
	if len(executionContext.MiddlewareStack) > currentIndex {
		currentMiddleware := executionContext.MiddlewareStack[currentIndex]
		currentMiddleware.Apply(run, func(newRun *models.PipelineRun) {
			executionContext.unwindStack(newRun, currentIndex+1)
		}, executionContext)
	}
}

func (executionContext *ExecutionContext) Execute(pipelineIdentifier string, writer io.Writer) {
	fullRun := executionContext.FullRun(WithIdentifier(&pipelineIdentifier))

	startProgress(executionContext, writer)
	fullRun.Close()
	fullRun.Wait()
	stopProgress(executionContext)

	outputResult(fullRun, writer)
	outputLogs(fullRun, writer)
	outputErrors(fullRun, writer)
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
	executionContext.Definitions = models.MergePipelineDefinitions(builtInPipelineDefinitions, userPipelineDefinitions)
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
		Stdout:    output,
	}
	return prompt.Run()
}
