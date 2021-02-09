// Package syncmiddleware provides a middleware to defer execution until a condition is fulfilled
package syncmiddleware

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"os"
	"time"
)

// Middleware is an execution synchronizer
type Middleware struct {
	LookupEnv func(string) (string, bool)
}

// String is a human-readable description
func (syncMiddleware Middleware) String() string {
	return "sync"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{
		LookupEnv: os.LookupEnv,
	}
}

type middlewareArguments struct {
	Pipes   []string
	EnvVars []string
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (syncMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := middlewareArguments{}
	pipeline.ParseArguments(&arguments, "sync", run)

	if arguments.Pipes != nil {
		for _, pipelineIdentifier := range arguments.Pipes {
			for _, dependentRun := range executionContext.Runs {
				if dependentRun.Identifier != nil && *dependentRun.Identifier == pipelineIdentifier {
					run.StartWaitGroup.Add(1)
					run.Log.Debug(
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
			run.Log.Debug(
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
