// Package catch provides a middleware for graceful handling of stderr output
package catch

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
)

// Middleware implements a handler for stderr output
type Middleware struct {
}

// String is a human-readable description
func (Middleware) String() string {
	return "catch"
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (catchMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	var argument pipeline.Reference = nil
	pipeline.ParseArguments(&argument, "catch", run)

	if argument != nil {

		next(run)

		var catchIdentifier *string = nil
		catchArguments := make(map[string]interface{}, 16)
		for pipelineIdentifier, pipelineArguments := range argument {
			catchIdentifier = pipelineIdentifier
			catchArguments = pipelineArguments
			break
		}

		run.Log.Trace(
			fields.DataStream(catchMiddleware, "creating stdout writer")...,
		)
		stdoutAppender := run.Stdout.WriteCloser()
		run.Log.Trace(
			fields.DataStream(catchMiddleware, "intercepting stderr")...,
		)
		stderrIntercept := run.Stderr.Intercept()
		go func() {
			// read the entire stderr output to enable multiline parsing
			errInput, stderrErr := ioutil.ReadAll(stderrIntercept)
			run.Log.PossibleError(stderrErr)

			if len(errInput) > 0 {
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(catchIdentifier),
					middleware.WithArguments(catchArguments),
					middleware.WithSetupFunc(func(errorRun *pipeline.Run) {
						run.Log.Trace(
							fields.DataStream(catchMiddleware, "merging parent stderr into child stdin")...,
						)
						errorRun.Stdin.MergeWith(bytes.NewReader(errInput))
					}),
					middleware.WithTearDownFunc(func(errorRun *pipeline.Run) {
						run.Log.Trace(
							fields.DataStream(catchMiddleware, "merging child stdout into parent stdout")...,
						)
						errorRun.Stdout.StartCopyingInto(stdoutAppender)
						run.Log.Trace(
							fields.DataStream(catchMiddleware, "replacing parent stderr with child stderr")...,
						)
						errorRun.Stderr.StartCopyingInto(stderrIntercept)
						executionContext.Connections = append(executionContext.Connections,
							pipeline.NewDataConnection(run, errorRun, "catch"))
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
