// Package _output provides a middleware to overwrite a pipe's output directly
package _output

import (
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"sync"
)

// Middleware is an output interceptor
type Middleware struct {
}

// String is a human-readable description
func (outputMiddleware Middleware) String() string {
	return "output"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type middlewareArguments struct {
	Process pipeline.Reference
	Text    *string
}

func newMiddlewareArguments() middlewareArguments {
	return middlewareArguments{
		Process: nil,
		Text:    nil,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (outputMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "output", run)

	next(run)

	if arguments.Text != nil {
		run.Log.Debug(
			fields.Symbol("↗️️"),
			fields.Message(*arguments.Text),
			fields.Middleware(outputMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := ioutil.ReadAll(stdoutIntercept)
			run.Log.PossibleError(err)
		}()
		go func() {
			_, err := stdoutIntercept.Write([]byte(*arguments.Text))
			run.Log.PossibleError(err)
			waitGroup.Wait()
			run.Log.PossibleError(stdoutIntercept.Close())
		}()
	}

	// switch is provided a list of regex patterns
	// it will use the first match to replace the output
	if len(arguments.Process) > 0 {
		run.Log.Debug(
			fields.Symbol("⍈"),
			fields.Message("process"),
			fields.Middleware(outputMiddleware),
		)

		run.Log.Trace(
			fields.DataStream(outputMiddleware, "intercepting stdout")...,
		)
		stdoutIntercept := run.Stdout.Intercept()
		run.Log.Trace(
			fields.DataStream(outputMiddleware, "creating stderr writer")...,
		)
		stderrAppender := run.Stderr.WriteCloser()
		parentLogWriter := run.Log.AddWriteCloserEntry()

		var childIdentifier *string
		childArguments := make(stringmap.StringMap, 10)
		for elseIdentifier, elseArguments := range arguments.Process {
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
					fields.DataStream(outputMiddleware, "merging parent's previous stdout into child stdin")...,
				)
				childRun.Stdin.MergeWith(stdoutIntercept)
			}),
			middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
				childRun.Log.Trace(
					fields.DataStream(outputMiddleware, "merging child stdout into parent's new stdout")...,
				)
				childRun.Stdout.StartCopyingInto(stdoutIntercept)
				childRun.Log.Trace(
					fields.DataStream(outputMiddleware, "merging child stderr into parent stderr")...,
				)
				childRun.Stderr.StartCopyingInto(stderrAppender)
				executionContext.AddConnection(run, childRun, "output processing")
				go func() {
					childRun.Wait()
					// need to clean up by closing the writers we created
					childRun.Log.PossibleError(stdoutIntercept.Close())
					childRun.Log.PossibleError(stderrAppender.Close())
				}()
			}))
	}
}
