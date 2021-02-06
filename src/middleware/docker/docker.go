// The `docker` middleware enables execution within a Docker (Compose) container
package docker

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Docker Executor
type DockerMiddleware struct {
}

func (_ DockerMiddleware) String() string {
	return "docker"
}

func NewDockerMiddleware() DockerMiddleware {
	return DockerMiddleware{}
}

func (dockerMiddleware DockerMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Service *string
	}{}
	middleware.ParseArgumentsIncludingParents(&arguments, "docker", run)

	if arguments.Service != nil {
		run.Log.DebugWithFields(
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
