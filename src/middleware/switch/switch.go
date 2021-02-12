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

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

type middlewareArguments = []struct {
	Pattern *string
	Text    *string
}

func newMiddlewareArguments() middlewareArguments {
	return middlewareArguments{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (switchMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "switch", run)

	next(run)

	if len(arguments) > 0 {
		run.Log.Debug(
			fields.Symbol("🔢"),
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

			foundMatch := false
			for _, switchStatement := range arguments {
				if switchStatement.Text == nil {
					continue
				}
				pattern := "anything"
				if switchStatement.Pattern != nil {
					pattern = *switchStatement.Pattern
					regex, err := regexp.Compile(*switchStatement.Pattern)
					run.Log.PossibleError(err)
					if err != nil || !regex.Match(inputData) {
						run.Log.Debug(
							fields.Symbol("🔢"),
							fields.Message("mismatch"),
							fields.Info(*switchStatement.Pattern),
							fields.Middleware("case"),
						)
						continue
					}
				}

				foundMatch = true
				run.Log.Debug(
					fields.Symbol("🔢"),
					fields.Message("match"),
					fields.Info(pattern),
					fields.Middleware("case"),
				)
				_, err := stdoutAppender.Write([]byte(*switchStatement.Text))
				run.Log.PossibleError(err)
				break
			}

			if !foundMatch {
				run.Log.Debug(
					fields.Symbol("🔢"),
					fields.Message("no match found"),
					fields.Middleware(switchMiddleware),
				)
			}

			run.Log.PossibleError(stdoutAppender.Close())
		}()
	}
}
