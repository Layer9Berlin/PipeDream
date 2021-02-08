// Package when provides a middleware that enables conditional execution
package when

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/evaluate"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
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

	shouldExecute, err := evaluate.EvaluateBool(argument)
	if err != nil {
		run.Log.Error(err)
		return
	}

	if shouldExecute {
		run.Log.DebugWithFields(
			fields.Symbol("?"),
			fields.Message("satisfied"),
			fields.Info(fmt.Sprintf("%q", argument)),
			fields.Middleware(whenMiddleware),
		)
		next(run)
	} else {
		run.Log.DebugWithFields(
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
