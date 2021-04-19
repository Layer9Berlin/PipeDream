// Package sync provides a middleware that makes the execution function wait until the run has completed
package sync

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
)

// Middleware is a conditional executor
type Middleware struct {
}

// String is a human-readable description
func (syncMiddleware Middleware) String() string {
	return "sync"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (syncMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	runSynchronously := false
	pipeline.ParseArgumentsIncludingParents(&runSynchronously, "sync", run)

	if runSynchronously {
		run.Log.Debug(
			fields.Symbol("="),
			fields.Message("synchronous run"),
			fields.Middleware(syncMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		stderrIntercept := run.Stdout.Intercept()
		next(run)
		go func() {
			completeStdout, _ := ioutil.ReadAll(stdoutIntercept)
			_, _ = stdoutIntercept.Write(completeStdout)
			_ = stdoutIntercept.Close()
		}()
		go func() {
			completeStderr, _ := ioutil.ReadAll(stderrIntercept)
			_, _ = stderrIntercept.Write(completeStderr)
			_ = stderrIntercept.Close()
		}()
		run.Wait()
		return
	}

	next(run)
}
