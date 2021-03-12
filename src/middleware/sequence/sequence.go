// Package sequence provides a middleware that executes pipes synchronously, one after the other
package sequence

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is a synchronous executor
type Middleware struct {
}

// String is a human-readable description
func (sequenceMiddleware Middleware) String() string {
	return "sequence"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type middlewareArguments = []pipeline.Reference

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (sequenceMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	middlewareArguments := middlewareArguments{}
	pipeline.ParseArguments(&middlewareArguments, "sequence", run)

	if len(middlewareArguments) > 0 {
		childIdentifiers, childArguments, info := pipeline.CollectReferences(middlewareArguments)
		run.Log.Trace(
			fields.Symbol("🔢"),
			fields.Info(info),
			fields.Middleware(sequenceMiddleware),
		)
		stdoutWriter := run.Stdout.WriteCloser()
		stderrWriter := run.Stderr.WriteCloser()
		var previousRun *pipeline.Run
		for index, childIdentifier := range childIdentifiers {
			childArguments := childArguments[index]
			// store the previous run to use in async block
			childPredecessor := previousRun
			childRun := executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(childIdentifier),
				middleware.WithArguments(childArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					if childPredecessor != nil {
						childRun.Log.Trace(
							fields.Symbol("🔢"),
							fields.Message("wait for completion"),
							fields.Info(previousRun.Name()),
							fields.Middleware(sequenceMiddleware),
						)
						childRun.StartWaitGroup.Add(1)
						go func() {
							childPredecessor.Wait()
							childRun.StartWaitGroup.Done()
						}()
					}
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					run.Log.Trace(
						fields.DataStream(sequenceMiddleware, "copy child stdout into parent stdout")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutWriter)
					run.Log.Trace(
						fields.DataStream(sequenceMiddleware, "copy child stderr into parent stderr")...,
					)
					childRun.Stderr.StartCopyingInto(stderrWriter)
					if previousRun != nil {
						executionContext.AddConnection(previousRun, childRun, "next")
					} else {
						executionContext.AddConnection(run, childRun, "sequence")
					}
				}))
			previousRun = childRun
		}
		go func() {
			previousRun.Wait()
			_ = stdoutWriter.Close()
			_ = stderrWriter.Close()
		}()
	}

	next(run)
}
