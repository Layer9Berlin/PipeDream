package inherit

import (
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
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
	run *models.PipelineRun,
	next func(*models.PipelineRun),
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
				_ = run.SetArgumentAtPath(parentValue, inheritedArgument)
				substitutions[inheritedArgument] = parentValue
			}
		}
		if len(substitutions) > 0 {
			run.Log.DebugWithFields(
				log_fields.Symbol("👪"),
				log_fields.Message("inherited arguments"),
				log_fields.Info(substitutions),
				log_fields.Middleware(inheritMiddleware),
			)
		}
	}
	next(run)
}
