// Package outputmiddleware provides a middleware to overwrite a pipe's output directly
package outputmiddleware

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"regexp"
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
func (outputMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
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
	if arguments.Switch != nil {
		run.Log.Debug(
			fields.Symbol("↗️️"),
			fields.Message("switch"),
			fields.Middleware(outputMiddleware),
		)

		// using the stdout intercept to be able to read and write stdout asynchronously
		stdoutIntercept := run.Stdout.Intercept()
		// do not close log yet, we may still want to write errors to it...
		run.LogClosingWaitGroup.Add(1)
		go func() {
			defer run.LogClosingWaitGroup.Done()
			// read the entire input data
			outputData, _ := ioutil.ReadAll(stdoutIntercept)

			foundMatch := false
			for _, switchStatement := range arguments.Switch {
				if switchStatement.Pattern == nil || switchStatement.Text == nil {
					continue
				}
				regex, err := regexp.Compile(*switchStatement.Pattern)
				run.Log.PossibleError(err)
				if err == nil && regex.Match(outputData) {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("match"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(outputMiddleware),
					)
					foundMatch = true
					_, err = stdoutIntercept.Write([]byte(*switchStatement.Text))
					run.Log.PossibleError(err)
					break
				} else {
					run.Log.Debug(
						fields.Symbol("↗️️"),
						fields.Message("mismatch"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(outputMiddleware),
					)
				}
			}

			if !foundMatch {
				if arguments.Else != nil {
					_, err := stdoutIntercept.Write([]byte(*arguments.Else))
					run.Log.PossibleError(err)
				}
			}

			run.Log.PossibleError(stdoutIntercept.Close())
		}()
	}
}
