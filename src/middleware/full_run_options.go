package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
)

// FullRunOption represents an option provided to an executions context's FullRun
type FullRunOption func(*FullRunOptions)

// WithIdentifier sets the identifier of the pipeline to be run
//
// It will be used to look up the matching definition, if any
func WithIdentifier(identifier *string) FullRunOption {
	return func(options *FullRunOptions) {
		options.pipelineIdentifier = identifier
	}
}

// WithParentRun sets the parent run, i.e. the enclosing run that triggered the full run in question
func WithParentRun(parentRun *pipeline.Run) FullRunOption {
	return func(options *FullRunOptions) {
		options.parentRun = parentRun
	}
}

// WithLogWriter sets an io.WriteCloser to which logs will be written
func WithLogWriter(logWriter io.WriteCloser) FullRunOption {
	return func(options *FullRunOptions) {
		options.logWriter = logWriter
	}
}

// WithArguments sets the full run's arguments
func WithArguments(arguments map[string]interface{}) FullRunOption {
	return func(options *FullRunOptions) {
		options.arguments = arguments
	}
}

// WithSetupFunc sets a function that will be executed between the full run's setup and main execution function
func WithSetupFunc(preCallback func(*pipeline.Run)) FullRunOption {
	return func(options *FullRunOptions) {
		options.preCallback = preCallback
	}
}

// WithTearDownFunc sets a function that will be executed after the full run's main execution function
func WithTearDownFunc(postCallback func(*pipeline.Run)) FullRunOption {
	return func(options *FullRunOptions) {
		options.postCallback = postCallback
	}
}

// FullRunOptions collects a number of different FullRunOption options into a single structure
type FullRunOptions struct {
	arguments          map[string]interface{}
	logWriter          io.WriteCloser
	parentRun          *pipeline.Run
	pipelineIdentifier *string
	postCallback       func(*pipeline.Run)
	preCallback        func(*pipeline.Run)
}
