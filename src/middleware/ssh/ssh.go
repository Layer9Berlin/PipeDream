package ssh

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
)

// SSH Executor
type SshMiddleware struct {
}

func (sshMiddleware SshMiddleware) String() string {
	return "ssh"
}

func NewSshMiddleware() SshMiddleware {
	return SshMiddleware{}
}

func (sshMiddleware SshMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Host *string
	}{}
	middleware.ParseArgumentsIncludingParents(&arguments, "ssh", run)

	if arguments.Host != nil {
		run.Log.DebugWithFields(
			log_fields.Symbol("👨‍💻"),
			log_fields.Message(arguments.Host),
			log_fields.Middleware(sshMiddleware),
		)
		prefixWithSshHost(run, *arguments.Host)
	}

	next(run)
}

func prefixWithSshHost(run *models.PipelineRun, sshHost string) {
	path := []string{"shell", "run"}
	// get the existing value - the shell -> run argument is not inheritable
	existingValue, err := run.ArgumentAtPath(path...)
	if err == nil {
		runArgumentAsString, runArgumentIsString := existingValue.(string)
		if runArgumentIsString {
			err := run.SetArgumentAtPath(fmt.Sprintf(`ssh %v %q`, sshHost, fmt.Sprintf("bash -l -c %q", runArgumentAsString)), path...)
			run.Log.PossibleError(err)
		}
	}
}
