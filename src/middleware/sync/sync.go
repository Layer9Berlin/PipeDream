package sync_middleware

import (
	"fmt"
	"os"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"time"
)

// Execution Synchronizer
type SyncMiddleware struct {
	LookupEnv func(string) (string, bool)
}

func (syncMiddleware SyncMiddleware) String() string {
	return "sync"
}

func NewSyncMiddleware() SyncMiddleware {
	return SyncMiddleware{
		LookupEnv: os.LookupEnv,
	}
}

type SyncMiddlewareArguments struct {
	waitFor []string
	envVars []string
}

func (syncMiddleware SyncMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := SyncMiddlewareArguments{}
	middleware.ParseArguments(&arguments, "sync", run)

	if arguments.waitFor != nil {
		for _, pipelineIdentifier := range arguments.waitFor {
			for _, dependentRun := range executionContext.Runs {
				if dependentRun.Identifier != nil && *dependentRun.Identifier == pipelineIdentifier {
					run.StartWaitGroup.Add(1)
					run.Log.DebugWithFields(
						log_fields.Symbol("🕙"),
						log_fields.Message(fmt.Sprintf("waiting for run %q", pipelineIdentifier)),
						log_fields.Middleware(syncMiddleware),
					)
					go func() {
						dependentRun.Wait()
						run.StartWaitGroup.Done()
					}()
				}
			}
		}
	}

	if arguments.envVars != nil {
		for _, envVar := range arguments.envVars {
			run.StartWaitGroup.Add(1)
			envVar := envVar
			run.Log.DebugWithFields(
				log_fields.Symbol("🕙"),
				log_fields.Message(fmt.Sprintf("waiting for env var %q to be set", envVar)),
				log_fields.Middleware(syncMiddleware),
			)
			go func() {
				for {
					if _, ok := syncMiddleware.LookupEnv(envVar); ok {
						break
					}
					time.Sleep(200)
				}
				run.StartWaitGroup.Done()
			}()
		}
	}

	next(run)
}
