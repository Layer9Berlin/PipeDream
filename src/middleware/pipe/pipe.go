package pipe

import (
	"bytes"
	"io"
	"io/ioutil"
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
		if run.Synchronous {
			run.Stdin.Close()
			run.Stdin.Wait()
			previousOutput = bytes.NewReader(run.Stdin.Bytes())
		} else {
			previousOutput = run.Stdin.Copy()
			go func() {
				run.Stdin.Close()
			}()
		}
		mergeDescription := "merging parent stdin into stdin"
		for index, childIdentifier := range childIdentifiers {
			identifier := childIdentifier
			arguments := childArguments[index]
			childRun := executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(arguments),
				middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
					childRun.Log.TraceWithFields(
						log_fields.DataStream(pipeMiddleware, mergeDescription)...,
					)
					childRun.Stdin.MergeWith(previousOutput)
				}),
				middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
					childRun.Log.TraceWithFields(
						log_fields.DataStream(pipeMiddleware, "copying stdout")...,
					)
					// write to the next run's input
					// or the parent's output, if this is the last child
					previousOutput = childRun.Stdout.Copy()
				}))
			mergeDescription = "merging previous stdout into stdin"
			if run.Synchronous {
				completeOutput, err := ioutil.ReadAll(previousOutput)
				run.Log.PossibleError(err)
				previousOutput = bytes.NewReader(completeOutput)
				childRun.Wait()
			}
		}
		run.Log.TraceWithFields(
			log_fields.DataStream(pipeMiddleware, "merging last child's stdout into stdout")...,
		)
		run.Stdout.MergeWith(previousOutput)
	} else {
		next(run)
	}
}
