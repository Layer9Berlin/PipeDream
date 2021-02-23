// Package when provides a middleware that enables conditional execution
package when

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/evaluate"
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is a conditional executor
type Middleware struct {
}

// String is a human-readable description
func (whenMiddleware Middleware) String() string {
	return "when"
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
func (whenMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	argument := ""
	pipeline.ParseArguments(&argument, "when", run)

	if argument == "" {
		next(run)
		return
	}

	shouldExecute, err := evaluate.Bool(argument)
	if err != nil {
		run.Log.Error(err)
		return
	}

	if shouldExecute {
		run.Log.Debug(
			fields.Symbol("?"),
			fields.Message("satisfied"),
			fields.Info(fmt.Sprintf("%q", argument)),
			fields.Middleware(whenMiddleware),
		)
		next(run)
	} else {
		run.Log.Debug(
			fields.Symbol("?"),
			fields.Message("not satisfied"),
			fields.Info(fmt.Sprintf("%q", argument)),
			fields.Color("lightgrey"),
			fields.Middleware(whenMiddleware),
		)
		arguments := pipeline.Reference{}
		pipeline.ParseArguments(&arguments, "else", run)

		if len(arguments) > 0 {
			run.Log.Debug(
				fields.Symbol("✖️"),
				fields.Info("else"),
				fields.Middleware(whenMiddleware),
			)
			run.Log.Trace(
				fields.DataStream(whenMiddleware, "copying stdin")...,
			)
			stdinCopy := run.Stdin.Copy()
			run.Log.Trace(
				fields.DataStream(whenMiddleware, "creating stdout writer")...,
			)
			stdoutAppender := run.Stdout.WriteCloser()
			run.Log.Trace(
				fields.DataStream(whenMiddleware, "creating stderr writer")...,
			)
			stderrAppender := run.Stderr.WriteCloser()
			// we return immediately and wait for the previous input to be available
			// then we execute a full run
			parentLogWriter := run.Log.AddWriteCloserEntry()

			var childIdentifier *string
			childArguments := make(stringmap.StringMap, 10)
			for elseIdentifier, elseArguments := range arguments {
				childIdentifier = elseIdentifier
				childArguments = elseArguments
				break
			}
			executionContext.FullRun(
				middleware.WithIdentifier(childIdentifier),
				middleware.WithParentRun(run),
				middleware.WithLogWriter(parentLogWriter),
				middleware.WithArguments(childArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					childRun.Log.Trace(
						fields.DataStream(whenMiddleware, "merging parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(stdinCopy)
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					childRun.Log.Trace(
						fields.DataStream(whenMiddleware, "merging child stdout into parent stdout")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutAppender)
					childRun.Log.Trace(
						fields.DataStream(whenMiddleware, "merging child stderr into parent stderr")...,
					)
					childRun.Stderr.StartCopyingInto(stderrAppender)
					executionContext.Connections = append(executionContext.Connections,
						pipeline.NewDataConnection(run, childRun, "else"))
					go func() {
						childRun.Wait()
						// need to clean up by closing the writers we created
						childRun.Log.PossibleError(stdoutAppender.Close())
						childRun.Log.PossibleError(stderrAppender.Close())
					}()
				}))
		} else {
			// the provided input should not be discarded, but passed through
			// (there might be a next invocation in the chain)
			run.Stdout.MergeWith(run.Stdin.Copy())
		}
	}
}
