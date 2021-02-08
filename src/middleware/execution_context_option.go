package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/parsing"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
)

type ExecutionContextOption func(*ExecutionContext)

func WithExecutionFunction(executionFunction func(run *pipeline.Run)) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.executionFunction = executionFunction
	}
}

func WithDefinitionsLookup(definitions pipeline.PipelineDefinitionsLookup) ExecutionContextOption {
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

func WithParser(parser *parsing.Parser) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.parser = parser
	}
}

func WithLogger(logger *logrus.Logger) ExecutionContextOption {
	return func(executionContext *ExecutionContext) {
		executionContext.Log = logger
	}
}
