package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
)

type FullRunOptions struct {
	arguments          map[string]interface{}
	logWriter          io.WriteCloser
	parentRun          *pipeline.Run
	pipelineIdentifier *string
	postCallback       func(*pipeline.Run)
	preCallback        func(*pipeline.Run)
}

type FullRunOption func(*FullRunOptions)

func WithIdentifier(identifier *string) FullRunOption {
	return func(options *FullRunOptions) {
		options.pipelineIdentifier = identifier
	}
}

func WithParentRun(parentRun *pipeline.Run) FullRunOption {
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

func WithSetupFunc(preCallback func(*pipeline.Run)) FullRunOption {
	return func(options *FullRunOptions) {
		options.preCallback = preCallback
	}
}

func WithTearDownFunc(postCallback func(*pipeline.Run)) FullRunOption {
	return func(options *FullRunOptions) {
		options.postCallback = postCallback
	}
}
