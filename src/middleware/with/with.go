// Package with provides a middleware that extracts and processes input patterns
package with

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"regexp"
	"sync"
)

// Middleware is a pattern extractor
type Middleware struct {
}

// String is a human-readable description
func (withMiddleware Middleware) String() string {
	return "with"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type withMiddlewareArguments struct {
	Pattern string
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (withMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	argument := withMiddlewareArguments{
		Pattern: "",
	}
	pipeline.ParseArguments(&argument, "with", run)

	next(run)

	if argument.Pattern != "" {
		run.Log.Debug(
			fields.Symbol("／"),
			fields.Middleware(withMiddleware),
			fields.Message("pattern"),
			fields.Info(argument.Pattern),
		)

		regex, err := regexp.Compile(argument.Pattern)
		if err != nil {
			run.Log.Error(err, fields.Middleware(withMiddleware))
			return
		}

		run.Log.Trace(
			fields.Symbol("⎇"),
			fields.Message("copying stdin"),
			fields.Middleware(withMiddleware),
		)
		stdinCopy := run.Stdin.Copy()
		run.Log.Trace(
			fields.Symbol("⎇"),
			fields.Message("intercepting stdout"),
			fields.Middleware(withMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		run.Log.Trace(
			fields.Symbol("⎇"),
			fields.Message("creating stderr writer"),
			fields.Middleware(withMiddleware),
		)
		stderrAppender := run.Stderr.WriteCloser()
		waitGroup := &sync.WaitGroup{}
		go func() {
			completeInput, err := ioutil.ReadAll(stdinCopy)
			inputMutex := &sync.Mutex{}
			run.Log.PossibleError(err)
			matches := regex.FindAllSubmatch(completeInput, -1)
			waitGroup.Add(len(matches))
			for _, match := range matches {
				match := match
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(run.Identifier),
					middleware.WithArguments(run.ArgumentsCopy()),
					middleware.WithSetupFunc(func(matchRun *pipeline.Run) {
						run.Log.Trace(
							fields.Symbol("⎇"),
							fields.Message("replacing stdin with regex match"),
							fields.Middleware(withMiddleware),
						)
						matchRun.Stdin.Replace(bytes.NewBuffer(match[0]))
					}),
					middleware.WithTearDownFunc(func(matchRun *pipeline.Run) {
						run.Log.Trace(
							fields.Symbol("⎇"),
							fields.Message("copying child stderr into parent stderr"),
							fields.Middleware(withMiddleware),
						)
						matchRun.Stderr.StartCopyingInto(stderrAppender)
						executionContext.AddConnection(run, matchRun, "with")
						go func() {
							defer waitGroup.Done()
							matchRun.Wait()
							defer inputMutex.Unlock()
							inputMutex.Lock()
							completeStdout := matchRun.Stdout.Bytes()
							completeInput = bytes.Replace(completeInput, match[0], completeStdout, -1)
						}()
					}))
			}
			go func() {
				waitGroup.Wait()
				_, err = stdoutIntercept.Write(completeInput)
				run.Log.PossibleError(err)
				run.Log.PossibleError(stdoutIntercept.Close())
				run.Log.PossibleError(stderrAppender.Close())
			}()
		}()
	}
}
