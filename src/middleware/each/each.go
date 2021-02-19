// Package each provides a middleware that copies some input into several child pipes running simultaneously
package each

import (
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"strings"
)

// Middleware is an input duplicator
type Middleware struct {
}

// String is a human-readable description
func (Middleware) String() string {
	return "each"
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
func (eachMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := make([]pipeline.Reference, 0, 10)
	pipeline.ParseArguments(&arguments, "each", run)

	haveChildren := len(arguments) > 0
	if haveChildren {
		childIdentifiers := make([]*string, 0, len(arguments))
		childArguments := make([]map[string]interface{}, 0, len(arguments))
		for _, childReference := range arguments {
			for pipelineIdentifier, pipelineArguments := range childReference {
				childIdentifiers = append(childIdentifiers, pipelineIdentifier)
				childArguments = append(childArguments, stringmap.CopyMap(pipelineArguments))
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
		run.Log.Debug(
			fields.Symbol("🔢"),
			fields.Message("each"),
			fields.Info(strings.Join(info, ", ")),
			fields.Middleware(eachMiddleware),
		)
		for index, childIdentifier := range childIdentifiers {
			arguments := childArguments[index]
			identifier := childIdentifier
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(identifier),
				middleware.WithArguments(arguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					run.Log.Trace(
						fields.DataStream(eachMiddleware, "copy parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(run.Stdin.Copy())
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					run.Log.Trace(
						fields.DataStream(eachMiddleware, "copy child stdout into parent stdout")...,
					)
					run.Stdout.MergeWith(childRun.Stdout.Copy())
					run.Log.Trace(
						fields.DataStream(eachMiddleware, "copy child stderr into parent stderr")...,
					)
					run.Stderr.MergeWith(childRun.Stderr.Copy())
					executionContext.Connections = append(executionContext.Connections,
						pipeline.NewDataConnection(run, childRun, "each"))
				}))
		}
	}

	next(run)
}
