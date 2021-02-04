package catch_each

import (
	"bufio"
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
	"sync"
)

// Line-Based Error Handler
type CatchEachMiddleware struct {
}

func (_ CatchEachMiddleware) String() string {
	return "catchEach"
}

func NewCatchEachMiddleware() CatchEachMiddleware {
	return CatchEachMiddleware{}
}

func (catchEachMiddleware CatchEachMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	argument := ""
	middleware.ParseArguments(&argument, "catchEach", run)

	if argument != "" {
		next(run)

		run.Log.TraceWithFields(
			log_fields.DataStream(catchEachMiddleware, "create stdout writer")...,
		)
		stdoutAppender := run.Stdout.WriteCloser()
		run.Log.TraceWithFields(
			log_fields.DataStream(catchEachMiddleware, "intercept stderr")...,
		)
		stderrIntercept := run.Stderr.Intercept()
		// write the shell command's errors to this pipe instead
		// we scan everything written to the pipe line by line
		scanner := bufio.NewScanner(stderrIntercept)
		go func() {
			waitGroup := &sync.WaitGroup{}
			for scanner.Scan() {
				waitGroup.Add(1)
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(&argument),
					middleware.WithSetupFunc(func(errorRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.DataStream(catchEachMiddleware, "merge regex match from parent stderr into child stdin")...,
						)
						errorRun.Stdin.MergeWith(bytes.NewReader(scanner.Bytes()))
					}),
					middleware.WithTearDownFunc(func(errorRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.DataStream(catchEachMiddleware, "merge child stdout into parent stdout")...,
						)
						errorRun.Stdout.StartCopyingInto(stdoutAppender)
						run.Log.TraceWithFields(
							log_fields.DataStream(catchEachMiddleware, "replace parent stderr with child stderr")...,
						)
						errorRun.Stderr.StartCopyingInto(stderrIntercept)
						go func() {
							defer waitGroup.Done()
							errorRun.Wait()
						}()
					}))
			}
			go func() {
				waitGroup.Wait()
				// close all writers once the child runs have completed
				run.Log.PossibleError(stdoutAppender.Close())
				run.Log.PossibleError(stderrIntercept.Close())
			}()
		}()
	} else {
		next(run)
	}

}
