package inputMiddleware

import (
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
	"io/ioutil"
)

// Input Interceptor
type InputMiddleware struct {
}

func (inputMiddleware InputMiddleware) String() string {
	return "input"
}

func NewInputMiddleware() InputMiddleware {
	return InputMiddleware{}
}

type InputMiddlewareArguments struct {
	Text *string
}

func NewOutputMiddlewareArguments() InputMiddlewareArguments {
	return InputMiddlewareArguments{
		Text: nil,
	}
}

func (inputMiddleware InputMiddleware) Apply(
	run *models.PipelineRun,
	next func(pipelineRun *models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := NewOutputMiddlewareArguments()
	middleware.ParseArguments(&arguments, "input", run)

	next(run)

	if arguments.Text != nil {
		run.Log.DebugWithFields(
			log_fields.Symbol("↘️"),
			log_fields.Message(*arguments.Text),
			log_fields.Middleware(inputMiddleware),
		)
		stdinIntercept := run.Stdin.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdinIntercept)
			run.Log.PossibleError(err)
		}()
		_, err := stdinIntercept.Write([]byte(*arguments.Text))
		run.Log.PossibleError(err)
		run.Log.PossibleError(stdinIntercept.Close())
	}
}
