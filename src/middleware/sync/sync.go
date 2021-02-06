// Package sync provides a middleware to defer execution until a condition is fulfilled
package syncmiddleware

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"os"
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
	Pipes   []string
	EnvVars []string
}

func (syncMiddleware SyncMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := SyncMiddlewareArguments{}
	middleware.ParseArguments(&arguments, "sync", run)

	if arguments.Pipes != nil {
		for _, pipelineIdentifier := range arguments.Pipes {
			for _, dependentRun := range executionContext.Runs {
				if dependentRun.Identifier != nil && *dependentRun.Identifier == pipelineIdentifier {
					run.StartWaitGroup.Add(1)
					run.Log.DebugWithFields(
						fields.Symbol("🕙"),
						fields.Message(fmt.Sprintf("waiting for run %q", pipelineIdentifier)),
						fields.Middleware(syncMiddleware),
					)
					go func() {
						dependentRun.Wait()
						run.StartWaitGroup.Done()
					}()
				}
			}
		}
	}

	if arguments.EnvVars != nil {
		for _, envVar := range arguments.EnvVars {
			run.StartWaitGroup.Add(1)
			envVar := envVar
			run.Log.DebugWithFields(
				fields.Symbol("🕙"),
				fields.Message(fmt.Sprintf("waiting for env var %q to be set", envVar)),
				fields.Middleware(syncMiddleware),
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
