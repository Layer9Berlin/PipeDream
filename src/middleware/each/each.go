package each

import (
	"bytes"
	"io"
	"pipedream/src/helpers/string_map"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"strings"
)

// Input Duplicator
type EachMiddleware struct {
}

func (_ EachMiddleware) String() string {
	return "each"
}

func NewEachMiddleware() EachMiddleware {
	return EachMiddleware{}
}

func (eachMiddleware EachMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := make([]middleware.PipelineReference, 0, 10)
	middleware.ParseArguments(&arguments, "each", run)

	next(run)

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
		for _, identifier := range childIdentifiers {
			if identifier == nil {
				info = append(info, "~")
			} else {
				info = append(info, *identifier)
			}
		}
		run.Log.DebugWithFields(
			log_fields.Symbol("🔢"),
			log_fields.Message("each"),
			log_fields.Info(strings.Join(info, ", ")),
			log_fields.Middleware(eachMiddleware),
		)
		inputReaders := make([]io.Reader, 0, len(childIdentifiers))
		if run.Synchronous {
			run.Stdin.Close()
			run.Stdin.Wait()
			for _, _ = range childIdentifiers {
				inputReaders = append(inputReaders, bytes.NewReader(run.Stdin.Bytes()))
			}
		} else {
			for _, _ = range childIdentifiers {
				inputReaders = append(inputReaders, run.Stdin.Copy())
			}
			go func() {
				run.Stdin.Close()
			}()
		}
		for index, childIdentifier := range childIdentifiers {
			arguments := childArguments[index]
			identifier := childIdentifier
			childRun := executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(arguments),
				middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
					run.Log.TraceWithFields(
						log_fields.DataStream(eachMiddleware, "copy parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(inputReaders[index])
				}),
				middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
					if !run.Synchronous {
						run.Log.TraceWithFields(
							log_fields.DataStream(eachMiddleware, "copy child stdout into parent stdout")...,
						)
						run.Stdout.MergeWith(childRun.Stdout.Copy())
						run.Log.TraceWithFields(
							log_fields.DataStream(eachMiddleware, "copy child stderr into parent stderr")...,
						)
						run.Stderr.MergeWith(childRun.Stderr.Copy())
					}
				}))
			if run.Synchronous {
				childRun.Wait()
				run.Stdout.MergeWith(bytes.NewReader(childRun.Stdout.Bytes()))
				run.Stderr.MergeWith(bytes.NewReader(childRun.Stderr.Bytes()))
			}
		}
	}
}
