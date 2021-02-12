// Package _input provides a middleware to overwrite a pipe's input directly
package _input

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"regexp"
	"sync"
)

// Middleware is an input interceptor
type Middleware struct {
}

// String is a human-readable description
func (inputMiddleware Middleware) String() string {
	return "input"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type middlewareArguments struct {
	Else   *string
	Switch []struct {
		Pattern *string
		Text    *string
	}
	Text *string
}

func newMiddlewareArguments() middlewareArguments {
	return middlewareArguments{
		Else:   nil,
		Switch: nil,
		Text:   nil,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (inputMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "input", run)

	next(run)

	if arguments.Text != nil {
		run.Log.Debug(
			fields.Symbol("↘️"),
			fields.Message(*arguments.Text),
			fields.Middleware(inputMiddleware),
		)
		stdinIntercept := run.Stdin.Intercept()
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := ioutil.ReadAll(stdinIntercept)
			run.Log.PossibleError(err)
		}()
		go func() {
			_, err := stdinIntercept.Write([]byte(*arguments.Text))
			run.Log.PossibleError(err)
			waitGroup.Wait()
			run.Log.PossibleError(stdinIntercept.Close())
		}()
	}

	// switch is provided a list of regex patterns
	// it will use the first match to replace the output
	if arguments.Switch != nil {
		run.Log.Debug(
			fields.Symbol("↘️"),
			fields.Message("switch"),
			fields.Middleware(inputMiddleware),
		)

		// using the stdout intercept to be able to read and write stdout asynchronously
		stdinCopy := run.Stdin.Copy()
		stdoutAppender := run.Stdout.WriteCloser()
		// do not close log yet, we may still want to write errors to it...
		run.LogClosingWaitGroup.Add(1)
		go func() {
			defer run.LogClosingWaitGroup.Done()
			// read the entire input data
			inputData, _ := ioutil.ReadAll(stdinCopy)

			foundMatch := false
			for _, switchStatement := range arguments.Switch {
				if switchStatement.Pattern == nil || switchStatement.Text == nil {
					continue
				}
				regex, err := regexp.Compile(*switchStatement.Pattern)
				run.Log.PossibleError(err)
				if err == nil && regex.Match(inputData) {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("match"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(inputMiddleware),
					)
					foundMatch = true
					_, err = stdoutAppender.Write([]byte(*switchStatement.Text))
					run.Log.PossibleError(err)
					break
				} else {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("mismatch"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(inputMiddleware),
					)
				}
			}

			if !foundMatch {
				if arguments.Else != nil {
					_, err := stdoutAppender.Write([]byte(*arguments.Else))
					run.Log.PossibleError(err)
				}
			}

			run.Log.PossibleError(stdoutAppender.Close())
		}()
	}
}
