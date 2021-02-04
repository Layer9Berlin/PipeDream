package docker

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
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
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Service *string
	}{}
	middleware.ParseArgumentsIncludingParents(&arguments, "docker", run)

	if arguments.Service != nil {
		run.Log.DebugWithFields(
			log_fields.Symbol("🐳"),
			log_fields.Message("docker-compose exec"),
			log_fields.Info(arguments.Service),
			log_fields.Middleware(dockerMiddleware),
		)
		prefixWithService(run, *arguments.Service)
	}

	next(run)
}

func prefixWithService(run *models.PipelineRun, service string) {
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
