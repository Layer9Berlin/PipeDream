// The `inherit` middleware passes arguments from a parent to its children
package inherit

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Arguments Propagator
type InheritMiddleware struct {
}

func (_ InheritMiddleware) String() string {
	return "inherit"
}

func NewInheritMiddleware() InheritMiddleware {
	return InheritMiddleware{}
}

func (inheritMiddleware InheritMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := make([]string, 0, 10)
	middleware.ParseArguments(&arguments, "inherit", run)

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
			run.Log.DebugWithFields(
				fields.Symbol("👪"),
				fields.Message("inherited arguments"),
				fields.Info(substitutions),
				fields.Middleware(inheritMiddleware),
			)
		}
	}
	next(run)
}
