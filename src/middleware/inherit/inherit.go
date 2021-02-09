// Package inherit provides a middleware that passes arguments from a parent to its children
package inherit

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is an arguments propagator
type Middleware struct {
}

// String is a human-readable description
func (Middleware) String() string {
	return "inherit"
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
func (inheritMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := make([]string, 0, 10)
	pipeline.ParseArguments(&arguments, "inherit", run)

	if run.Parent != nil {
		substitutions := make(map[string]interface{}, len(arguments))
		pipelineArguments := run.ArgumentsCopy()
		parentArguments := run.Parent.ArgumentsCopy()
		for _, inheritedArgument := range arguments {
			parentValue, haveParentValue := parentArguments[inheritedArgument]
			_, haveExistingValue := pipelineArguments[inheritedArgument]
			if !haveExistingValue && haveParentValue {
				err := run.SetArgumentAtPath(parentValue, inheritedArgument)
				run.Log.PossibleError(err)
				substitutions[inheritedArgument] = parentValue
			}
		}
		if len(substitutions) > 0 {
			run.Log.Debug(
				fields.Symbol("👪"),
				fields.Message("inherited arguments"),
				fields.Info(substitutions),
				fields.Middleware(inheritMiddleware),
			)
		}
	}
	next(run)
}
