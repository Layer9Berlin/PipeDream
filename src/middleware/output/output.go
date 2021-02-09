// Package outputmiddleware provides a middleware to overwrite a pipe's output directly
package outputmiddleware

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
)

// Middleware is an output interceptor
type Middleware struct {
}

// String is a human-readable description
func (outputMiddleware Middleware) String() string {
	return "output"
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
func (outputMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "output", run)

	next(run)

	if arguments.Text != nil {
		run.Log.Debug(
			fields.Symbol("↗️️"),
			fields.Message(*arguments.Text),
			fields.Middleware(outputMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			run.Log.PossibleError(err)
		}()
		go func() {
			_, err := stdoutIntercept.Write([]byte(*arguments.Text))
			run.Log.PossibleError(err)
			run.Log.PossibleError(stdoutIntercept.Close())
		}()
	}
}
