// Package docker provides a middleware for execution within a Docker (Compose) container
package docker

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is a Docker (Compose) executor
type Middleware struct {
}

// String is a human-readable description
func (Middleware) String() string {
	return "docker"
}

// NewMiddleware create a new Middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (dockerMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Service *string
	}{}
	pipeline.ParseArgumentsIncludingParents(&arguments, "docker", run)

	if arguments.Service != nil {
		run.Log.Debug(
			fields.Symbol("🐳"),
			fields.Message("docker-compose exec"),
			fields.Info(arguments.Service),
			fields.Middleware(dockerMiddleware),
		)
		prefixWithService(run, *arguments.Service)
	}

	next(run)
}

func prefixWithService(run *pipeline.Run, service string) {
	path := []string{"shell", "run"}
	// get the existing value - the shell -> run argument is not inheritable
	existingValue, err := run.ArgumentAtPath(path...)
	if err == nil {
		existingValueAsString, existingValueIsString := existingValue.(string)
		if existingValueIsString {
			err := run.SetArgumentAtPath(fmt.Sprintf("docker-compose exec -T %v %v", service, existingValueAsString), path...)
			run.Log.PossibleError(err)
		}
	}
}
