package catch

import (
	"bytes"
	"io/ioutil"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
)

// Error Handler
type CatchMiddleware struct {
}

func (_ CatchMiddleware) String() string {
	return "catch"
}

func NewCatchMiddleware() CatchMiddleware {
	return CatchMiddleware{}
}

func (catchMiddleware CatchMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	argument := ""
	middleware.ParseArguments(&argument, "catch", run)

	if argument != "" {

		next(run)

		run.Log.TraceWithFields(
			log_fields.DataStream(catchMiddleware, "creating stdout writer")...,
		)
		stdoutAppender := run.Stdout.WriteCloser()
		run.Log.TraceWithFields(
			log_fields.DataStream(catchMiddleware, "intercepting stderr")...,
		)
		stderrIntercept := run.Stderr.Intercept()
		go func() {
			// read the entire stderr output to enable multiline parsing
			errInput, stderrErr := ioutil.ReadAll(stderrIntercept)
			run.Log.PossibleError(stderrErr)

			if len(errInput) > 0 {
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(&argument),
					middleware.WithSetupFunc(func(errorRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.DataStream(catchMiddleware, "merging parent stderr into child stdin")...,
						)
						errorRun.Stdin.MergeWith(bytes.NewReader(errInput))
					}),
					middleware.WithTearDownFunc(func(errorRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.DataStream(catchMiddleware, "merging child stdout into parent stdout")...,
						)
						errorRun.Stdout.StartCopyingInto(stdoutAppender)
						run.Log.TraceWithFields(
							log_fields.DataStream(catchMiddleware, "replacing parent stderr with child stderr")...,
						)
						errorRun.Stderr.StartCopyingInto(stderrIntercept)
						go func() {
							errorRun.Wait()
							// need to clean up by closing the writers we created
							run.Log.PossibleError(stdoutAppender.Close())
							run.Log.PossibleError(stderrIntercept.Close())
						}()
					}))
			} else {
				// need to clean up by closing the writers we created
				run.Log.PossibleError(stdoutAppender.Close())
				run.Log.PossibleError(stderrIntercept.Close())
			}
		}()
	} else {
		next(run)
	}
}
