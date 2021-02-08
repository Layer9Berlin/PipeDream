// Package ssh provides a middleware enabling remote command execution via SSH
package ssh

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
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
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Host *string
	}{}
	pipeline.ParseArgumentsIncludingParents(&arguments, "ssh", run)

	if arguments.Host != nil {
		run.Log.DebugWithFields(
			fields.Symbol("👨‍💻"),
			fields.Message(arguments.Host),
			fields.Middleware(sshMiddleware),
		)
		prefixWithSshHost(run, *arguments.Host)
	}

	next(run)
}

func prefixWithSshHost(run *pipeline.Run, sshHost string) {
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
