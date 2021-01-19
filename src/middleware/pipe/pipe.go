package pipe

import (
	"io"
	"pipedream/src/helpers/string_map"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"strings"
)

// Invocation Chain
type PipeMiddleware struct{}

func (pipeMiddleware PipeMiddleware) String() string {
	return "pipe"
}

func NewPipeMiddleware() PipeMiddleware {
	return PipeMiddleware{}
}

func (pipeMiddleware PipeMiddleware) Apply(
	run *models.PipelineRun,
	next func(pipelineRun *models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := make([]middleware.PipelineReference, 0, 10)
	middleware.ParseArguments(&arguments, "pipe", run)

	haveChildren := len(arguments) > 0
	if haveChildren {
		childIdentifiers := make([]*string, 0, len(arguments))
		childArguments := make([]map[string]interface{}, 0, len(arguments))
		for _, childReference := range arguments {
			for pipelineIdentifier, pipelineArguments := range childReference {
				childIdentifiers = append(childIdentifiers, pipelineIdentifier)
				childArguments = append(childArguments, string_map.CopyMap(pipelineArguments))
			}
		}

		info := make([]string, 0, len(childIdentifiers))
		for _, childIdentifier := range childIdentifiers {
			if childIdentifier == nil {
				info = append(info, "anonymous")
			} else {
				info = append(info, *childIdentifier)
			}
		}

		switch len(info) {
		case 0:
			run.Log.DebugWithFields(
				log_fields.Symbol("⇣"),
				log_fields.Message("no invocation"),
				log_fields.Middleware(pipeMiddleware),
			)
			next(run)
			return
		case 1:
			run.Log.DebugWithFields(
				log_fields.Symbol("⇣"),
				log_fields.Message("single invocation: "+strings.Join(info, ", ")),
				log_fields.Middleware(pipeMiddleware),
			)
		default:
			run.Log.DebugWithFields(
				log_fields.Symbol("⇣"),
				log_fields.Message("invocation chain: "+strings.Join(info, ", ")),
				log_fields.Middleware(pipeMiddleware),
			)
		}
		var previousOutput io.Reader = nil
		for index, childIdentifier := range childIdentifiers {
			identifier := childIdentifier
			arguments := childArguments[index]
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(arguments),
				middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
					if index == 0 {
						childRun.Log.TraceWithFields(
							log_fields.DataStream(pipeMiddleware, "merging parent stdin into stdin")...,
						)
						childRun.Stdin.MergeWith(run.Stdin.Copy())
					} else {
						childRun.Log.TraceWithFields(
							log_fields.DataStream(pipeMiddleware, "merging previous stdout into stdin")...,
						)
						childRun.Stdin.MergeWith(previousOutput)
					}
				}),
				middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
					childRun.Log.TraceWithFields(
						log_fields.DataStream(pipeMiddleware, "copying stdout")...,
					)
					// write to the next run's input
					// or the parent's output, if this is the last child
					previousOutput = childRun.Stdout.Copy()
				}))
		}
		run.Log.TraceWithFields(
			log_fields.DataStream(pipeMiddleware, "merging last child's stdout into stdout")...,
		)
		run.Stdout.MergeWith(previousOutput)
	}

	next(run)
}
