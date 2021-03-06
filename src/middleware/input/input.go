// Package _input provides a middleware to overwrite a pipe's input directly
package _input

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"sync"
)

// Middleware is an input interceptor
type Middleware struct {
}

// String is a human-readable description
func (inputMiddleware Middleware) String() string {
	return "input"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type middlewareArguments struct {
	Text *string
}

func newMiddlewareArguments() middlewareArguments {
	return middlewareArguments{
		Text: nil,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (inputMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "input", run)

	next(run)

	if arguments.Text != nil {
		run.Log.Debug(
			fields.Symbol("↘️"),
			fields.Message(*arguments.Text),
			fields.Middleware(inputMiddleware),
		)
		stdinIntercept := run.Stdin.Intercept()
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := ioutil.ReadAll(stdinIntercept)
			run.Log.PossibleError(err)
		}()
		go func() {
			_, err := stdinIntercept.Write([]byte(*arguments.Text))
			run.Log.PossibleError(err)
			waitGroup.Wait()
			run.Log.PossibleError(stdinIntercept.Close())
		}()
	}
}
