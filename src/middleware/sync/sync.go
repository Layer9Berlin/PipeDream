package sync_middleware

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

// Synchronous Execution Chain
type SyncMiddleware struct {
}

func (syncMiddleware SyncMiddleware) String() string {
	return "sync"
}

func NewSyncMiddleware() SyncMiddleware {
	return SyncMiddleware{}
}

func (syncMiddleware SyncMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := make([]middleware.PipelineReference, 0, 10)
	middleware.ParseArguments(&arguments, "sync", run)

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
		case 1:
			run.Log.DebugWithFields(
				log_fields.Symbol("⇣"),
				log_fields.Message("single invocation: "+strings.Join(info, ", ")),
				log_fields.Middleware(syncMiddleware),
			)
		default:
			run.Log.DebugWithFields(
				log_fields.Symbol("⇣"),
				log_fields.Message("invocation chain: "+strings.Join(info, ", ")),
				log_fields.Middleware(syncMiddleware),
			)
		}

		initialStdin := run.Stdin.Copy()
		finalResultWriteCloser := run.Stdout.WriteCloser()
		startFullRunForEachChild(childIdentifiers, run, childArguments, initialStdin, finalResultWriteCloser, executionContext, 0)
	}

	next(run)
}

func startFullRunForEachChild(
	childIdentifiers []*string,
	run *models.PipelineRun,
	arguments []map[string]interface{},
	stdinReader io.Reader,
	finalResultWriteCloser io.WriteCloser,
	executionContext *middleware.ExecutionContext,
	index int,
) {
	if len(childIdentifiers) == 0 {
		go func() {
			_, _ = io.Copy(finalResultWriteCloser, stdinReader)
			_ = finalResultWriteCloser.Close()
		}()
		return
	}
	firstChildIdentifier, childIdentifiers := childIdentifiers[0], childIdentifiers[1:]
	executionContext.FullRun(
		middleware.WithParentRun(run),
		middleware.WithIdentifier(firstChildIdentifier),
		middleware.WithArguments(arguments[index]),
		middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
			childRun.Stdin.MergeWith(stdinReader)
		}),
		middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
			childStdout := childRun.Stdout.Copy()
			// return immediately, but wait for execution to complete and call next child
			go func() {
				completeOutput, _ := ioutil.ReadAll(childStdout)
				childRun.Wait()
				startFullRunForEachChild(childIdentifiers, run, arguments, bytes.NewReader(completeOutput), finalResultWriteCloser, executionContext, index+1)
			}()
		}),
	)
}
