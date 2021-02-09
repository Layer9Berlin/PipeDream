package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/parsing"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"io"
)

// ExecutionContextOption is an option that can be applied to an ExecutionContext or provided to the NewExecutionContext constructor
type ExecutionContextOption func(*ExecutionContext)

// WithDefinitionsLookup sets the execution context's definitions lookup
func WithDefinitionsLookup(definitions pipeline.DefinitionsLookup) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.Definitions = definitions
	}
}

// WithProjectPath sets the execution context's project path
func WithProjectPath(projectPath string) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.ProjectPath = projectPath
	}
}

// WithMiddlewareStack sets the execution context's middleware stack
func WithMiddlewareStack(stack []Middleware) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.MiddlewareStack = stack
	}
}

// WithParser sets the execution context's parser
func WithParser(parser *parsing.Parser) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.parser = parser
	}
}

// WithLogger sets the execution context's logger
func WithLogger(logger *logrus.Logger) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.Log = logger
	}
}

// WithExecutionFunction set the execution context's execution function
//
// You don't need to replace the standard implementation, which unwinds a middleware stack,
// unless you want to specify a specific behaviour e.g. for tests.
func WithExecutionFunction(executionFunction func(run *pipeline.Run)) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.executionFunction = executionFunction
	}
}

// WithUserPromptImplementation sets the execution context's implementation of a user prompt
//
// By default, this will use promptui to show an interactive prompt to the user,
// but you may want to override it for tests.
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
