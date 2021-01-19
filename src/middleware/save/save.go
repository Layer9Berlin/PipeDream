package save

import (
	"fmt"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"os"
)

// Env Var Storer
type SaveMiddleware struct {
	valueSetter func(string, string) error
}

func (saveMiddleware SaveMiddleware) String() string {
	return "save"
}

func NewSaveMiddleware() SaveMiddleware {
	return NewSaveMiddlewareWithValueSetter(os.Setenv)
}

func NewSaveMiddlewareWithValueSetter(valueSetter func(string, string) error) SaveMiddleware {
	return SaveMiddleware{
		valueSetter: valueSetter,
	}
}

func (saveMiddleware SaveMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	var envVarName *string = nil
	middleware.ParseArguments(&envVarName, "save", run)

	if envVarName != nil {
		run.Log.DebugWithFields(
			log_fields.Symbol("💾"),
			log_fields.Message("waiting for complete output"),
			log_fields.Info(*envVarName),
			log_fields.Middleware(saveMiddleware),
		)
		run.LogClosingWaitGroup.Add(1)
		go func() {
			defer run.LogClosingWaitGroup.Done()
			run.Stdout.Wait()
			run.Log.DebugWithFields(
				log_fields.Symbol("💾"),
				log_fields.Message(fmt.Sprintf("saving %v bytes of output", run.Stdout.Len())),
				log_fields.Info(*envVarName),
				log_fields.Middleware(saveMiddleware),
			)
			run.Log.PossibleError(saveMiddleware.valueSetter(*envVarName, run.Stdout.String()))
		}()
	}

	next(run)
}

type SaveMiddlewareEntry struct {
	path []string
	root *models.PipelineRun
}
