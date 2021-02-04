package when

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/helpers/custom_evaluate"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
)

// Conditional Executor
type WhenMiddleware struct {
}

func (whenMiddleware WhenMiddleware) String() string {
	return "when"
}

func NewWhenMiddleware() WhenMiddleware {
	return WhenMiddleware{}
}

func (whenMiddleware WhenMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	argument := ""
	middleware.ParseArguments(&argument, "when", run)

	if argument == "" {
		next(run)
		return
	}

	shouldExecute, err := custom_evaluate.EvaluateBool(argument)
	if err != nil {
		run.Log.Error(err)
		return
	}

	if shouldExecute {
		run.Log.DebugWithFields(
			log_fields.Symbol("?"),
			log_fields.Message("satisfied"),
			log_fields.Info(fmt.Sprintf("%q", argument)),
			log_fields.Middleware(whenMiddleware),
		)
		next(run)
	} else {
		run.Log.DebugWithFields(
			log_fields.Symbol("?"),
			log_fields.Message("not satisfied"),
			log_fields.Info(fmt.Sprintf("%q", argument)),
			log_fields.Color("lightgrey"),
			log_fields.Middleware(whenMiddleware),
		)
		// the provided input should not be discarded, but passed through
		// (there might be a next invocation in the chain)
		run.Stdout.MergeWith(run.Stdin.Copy())
	}
}
