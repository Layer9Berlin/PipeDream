// Package when provides a middleware that enables conditional execution
package when

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/evaluate"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is a conditional executor
type Middleware struct {
}

// String is a human-readable description
func (whenMiddleware Middleware) String() string {
	return "when"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (whenMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	argument := ""
	pipeline.ParseArguments(&argument, "when", run)

	if argument == "" {
		next(run)
		return
	}

	shouldExecute, err := evaluate.Bool(argument)
	if err != nil {
		run.Log.Error(err)
		return
	}

	if shouldExecute {
		run.Log.Debug(
			fields.Symbol("?"),
			fields.Message("satisfied"),
			fields.Info(fmt.Sprintf("%q", argument)),
			fields.Middleware(whenMiddleware),
		)
		next(run)
	} else {
		run.Log.Debug(
			fields.Symbol("?"),
			fields.Message("not satisfied"),
			fields.Info(fmt.Sprintf("%q", argument)),
			fields.Color("lightgrey"),
			fields.Middleware(whenMiddleware),
		)
		// the provided input should not be discarded, but passed through
		// (there might be a next invocation in the chain)
		run.Stdout.MergeWith(run.Stdin.Copy())
	}
}
