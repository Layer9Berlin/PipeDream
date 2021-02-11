// Package catcheach provides a middleware for graceful handling of stderr output based on regex patterns
package catch_each

import (
	"bufio"
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"sync"
)

// Middleware is a line-based error handler
type Middleware struct {
}

// String is a human-readable description
func (Middleware) String() string {
	return "catchEach"
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
func (catchEachMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	argument := ""
	pipeline.ParseArguments(&argument, "catchEach", run)

	if argument != "" {
		next(run)

		run.Log.Trace(
			fields.DataStream(catchEachMiddleware, "create stdout writer")...,
		)
		stdoutAppender := run.Stdout.WriteCloser()
		run.Log.Trace(
			fields.DataStream(catchEachMiddleware, "intercept stderr")...,
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
					middleware.WithSetupFunc(func(errorRun *pipeline.Run) {
						run.Log.Trace(
							fields.DataStream(catchEachMiddleware, "merge regex match from parent stderr into child stdin")...,
						)
						errorRun.Stdin.MergeWith(bytes.NewReader(scanner.Bytes()))
					}),
					middleware.WithTearDownFunc(func(errorRun *pipeline.Run) {
						run.Log.Trace(
							fields.DataStream(catchEachMiddleware, "merge child stdout into parent stdout")...,
						)
						errorRun.Stdout.StartCopyingInto(stdoutAppender)
						run.Log.Trace(
							fields.DataStream(catchEachMiddleware, "replace parent stderr with child stderr")...,
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
