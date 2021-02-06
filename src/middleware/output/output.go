// The `output` middleware overwrites a pipe's output directly
package outputmiddleware

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io/ioutil"
)

// Output Interceptor
type OutputMiddleware struct {
}

func (outputMiddleware OutputMiddleware) String() string {
	return "output"
}

func NewOutputMiddleware() OutputMiddleware {
	return OutputMiddleware{}
}

type OutputMiddlewareArguments struct {
	Text *string
}

func NewOutputMiddlewareArguments() OutputMiddlewareArguments {
	return OutputMiddlewareArguments{
		Text: nil,
	}
}

func (outputMiddleware OutputMiddleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := NewOutputMiddlewareArguments()
	middleware.ParseArguments(&arguments, "output", run)

	next(run)

	if arguments.Text != nil {
		run.Log.DebugWithFields(
			fields.Symbol("↗️️"),
			fields.Message(*arguments.Text),
			fields.Middleware(outputMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			run.Log.PossibleError(err)
		}()
		go func() {
			_, err := stdoutIntercept.Write([]byte(*arguments.Text))
			run.Log.PossibleError(err)
			run.Log.PossibleError(stdoutIntercept.Close())
		}()
	}
}
