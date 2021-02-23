// Package _switch provides a middleware that replaces the output using the first match from a list of regexes
package _switch

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
	"regexp"
)

// Middleware is a pattern matcher
type Middleware struct {
}

// String is a human-readable description
func (switchMiddleware Middleware) String() string {
	return "switch"
}

type middlewareArguments = []struct {
	Pattern *string
	Text    *string
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
func (switchMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := middlewareArguments{}
	pipeline.ParseArguments(&arguments, "switch", run)

	// switch is provided a list of regex patterns
	// it will use the first match to replace the output
	if len(arguments) > 0 {
		run.Log.Debug(
			fields.Symbol("☞️"),
			fields.Middleware(switchMiddleware),
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

			for _, switchStatement := range arguments {
				if switchStatement.Text == nil {
					continue
				}
				if switchStatement.Pattern == nil {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("match"),
						fields.Info("default"),
						fields.Middleware(switchMiddleware),
					)
					_, err := stdoutAppender.Write([]byte(*switchStatement.Text))
					run.Log.PossibleError(err)
					break
				}
				regex, err := regexp.Compile("(?m)" + *switchStatement.Pattern)
				run.Log.PossibleError(err)
				if err == nil && regex.Match(inputData) {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("match"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(switchMiddleware),
					)
					_, err = stdoutAppender.Write([]byte(*switchStatement.Text))
					run.Log.PossibleError(err)
					break
				} else {
					run.Log.Debug(
						fields.Symbol("🔢"),
						fields.Message("mismatch"),
						fields.Info(*switchStatement.Pattern),
						fields.Middleware(switchMiddleware),
					)
				}
			}

			run.Log.PossibleError(stdoutAppender.Close())
		}()
	}

	next(run)
}
