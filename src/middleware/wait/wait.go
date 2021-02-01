package wait

import (
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
)

// Execution Synchronizer
type WaitMiddleware struct {
}

func (waitMiddleware WaitMiddleware) String() string {
	return "wait"
}

func NewWaitMiddleware() WaitMiddleware {
	return WaitMiddleware{}
}

func (waitMiddleware WaitMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	waitArgument := false
	middleware.ParseArguments(&waitArgument, "wait", run)
	if waitArgument {
		run.Log.DebugWithFields(
			log_fields.Symbol("💤"),
			log_fields.Message("waiting..."),
			log_fields.Middleware(waitMiddleware),
		)
		run.Synchronous = true
	}
	next(run)
}
