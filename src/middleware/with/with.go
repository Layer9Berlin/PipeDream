// The `with` middleware extracts and processes input patterns
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

// Pattern Extractor
type WithMiddleware struct {
}

func (withMiddleware WithMiddleware) String() string {
	return "with"
}

func NewWithMiddleware() WithMiddleware {
	return WithMiddleware{}
}

type withMiddlewareArguments struct {
	Pattern string
}

func (withMiddleware WithMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	argument := withMiddlewareArguments{
		Pattern: "",
	}
	middleware.ParseArguments(&argument, "with", run)

	next(run)

	if argument.Pattern != "" {
		run.Log.DebugWithFields(
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

		run.Log.TraceWithFields(
			fields.Symbol("⎇"),
			fields.Message("copying stdin"),
			fields.Middleware(withMiddleware),
		)
		stdinCopy := run.Stdin.Copy()
		run.Log.TraceWithFields(
			fields.Symbol("⎇"),
			fields.Message("intercepting stdout"),
			fields.Middleware(withMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		run.Log.TraceWithFields(
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
						run.Log.TraceWithFields(
							fields.Symbol("⎇"),
							fields.Message("replacing stdin with regex match"),
							fields.Middleware(withMiddleware),
						)
						matchRun.Stdin.Replace(bytes.NewBuffer(match[0]))
					}),
					middleware.WithTearDownFunc(func(matchRun *pipeline.Run) {
						run.Log.TraceWithFields(
							fields.Symbol("⎇"),
							fields.Message("copying child stderr into parent stderr"),
							fields.Middleware(withMiddleware),
						)
						matchRun.Stderr.StartCopyingInto(stderrAppender)

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
