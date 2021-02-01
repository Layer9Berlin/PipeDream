package outputMiddleware

import (
	"io/ioutil"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
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
	run *models.PipelineRun,
	next func(pipelineRun *models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := NewOutputMiddlewareArguments()
	middleware.ParseArguments(&arguments, "output", run)

	next(run)

	if arguments.Text != nil {
		run.Log.DebugWithFields(
			log_fields.Symbol("↗️️"),
			log_fields.Message(*arguments.Text),
			log_fields.Middleware(outputMiddleware),
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
