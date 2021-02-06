// The `pipe` middleware executes several commands in sequence
package pipe

import (
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
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
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
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
				childArguments = append(childArguments, stringmap.CopyMap(pipelineArguments))
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
				fields.Symbol("⇣"),
				fields.Message("no invocation"),
				fields.Middleware(pipeMiddleware),
			)
			next(run)
			return
		case 1:
			run.Log.DebugWithFields(
				fields.Symbol("⇣"),
				fields.Message("single invocation: "+strings.Join(info, ", ")),
				fields.Middleware(pipeMiddleware),
			)
		default:
			run.Log.DebugWithFields(
				fields.Symbol("⇣"),
				fields.Message("invocation chain: "+strings.Join(info, ", ")),
				fields.Middleware(pipeMiddleware),
			)
		}
		var previousOutput io.Reader = nil
		for index, childIdentifier := range childIdentifiers {
			identifier := childIdentifier
			childRunArguments := childArguments[index]
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(childRunArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					if index == 0 {
						childRun.Log.TraceWithFields(
							fields.DataStream(pipeMiddleware, "merging parent stdin into stdin")...,
						)
						childRun.Stdin.MergeWith(run.Stdin.Copy())
					} else {
						childRun.Log.TraceWithFields(
							fields.DataStream(pipeMiddleware, "merging previous stdout into stdin")...,
						)
						childRun.Stdin.MergeWith(previousOutput)
					}
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					childRun.Log.TraceWithFields(
						fields.DataStream(pipeMiddleware, "copying stdout")...,
					)
					// write to the next run's input
					// or the parent's output, if this is the last child
					previousOutput = childRun.Stdout.Copy()
				}))
		}
		run.Log.TraceWithFields(
			fields.DataStream(pipeMiddleware, "merging last child's stdout into stdout")...,
		)
		run.Stdout.MergeWith(previousOutput)
	}

	next(run)
}
