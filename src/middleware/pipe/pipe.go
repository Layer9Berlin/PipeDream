// Package pipe provides a middleware to execute several commands in sequence
package pipe

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
	"strings"
)

// Middleware is an invocation chainer
type Middleware struct{}

func (pipeMiddleware Middleware) String() string {
	return "pipe"
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
func (pipeMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := make([]pipeline.Reference, 0, 10)
	pipeline.ParseArguments(&arguments, "pipe", run)

	haveChildren := len(arguments) > 0
	if haveChildren {
		childIdentifiers, childArguments, info := pipeline.CollectReferences(arguments)

		switch len(info) {
		case 0:
			run.Log.Debug(
				fields.Symbol("⇣"),
				fields.Message("no invocation"),
				fields.Middleware(pipeMiddleware),
			)
			next(run)
			return
		case 1:
			run.Log.Debug(
				fields.Symbol("⇣"),
				fields.Message("single invocation: "+strings.Join(info, ", ")),
				fields.Middleware(pipeMiddleware),
			)
		default:
			run.Log.Debug(
				fields.Symbol("⇣"),
				fields.Message("invocation chain: "+strings.Join(info, ", ")),
				fields.Middleware(pipeMiddleware),
			)
		}
		var previousOutput io.Reader = nil
		previousRun := run
		for index, childIdentifier := range childIdentifiers {
			identifier := childIdentifier
			childRunArguments := childArguments[index]
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(childRunArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					if index == 0 {
						childRun.Log.Trace(
							fields.DataStream(pipeMiddleware, "merging parent stdin into stdin")...,
						)
						childRun.Stdin.MergeWith(run.Stdin.Copy())
					} else {
						childRun.Log.Trace(
							fields.DataStream(pipeMiddleware, "merging previous stdout into stdin")...,
						)
						childRun.Stdin.MergeWith(previousOutput)
					}
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					executionContext.AddConnection(previousRun, childRun, "pipe")
					childRun.Log.Trace(
						fields.DataStream(pipeMiddleware, "copying stdout")...,
					)
					// write to the next run's input
					// or the parent's output, if this is the last child
					previousOutput = childRun.Stdout.Copy()
					previousRun = childRun
				}))
		}
		run.Log.Trace(
			fields.DataStream(pipeMiddleware, "merging last child's stdout into stdout")...,
		)
		run.Stdout.MergeWith(previousOutput)
	}

	next(run)
}
