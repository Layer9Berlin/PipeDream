// Package ssh provides a middleware enabling remote command execution via SSH
package ssh

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware is an SSH executor
type Middleware struct {
}

// String is a human-readable description
func (sshMiddleware Middleware) String() string {
	return "ssh"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (sshMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := struct {
		Host *string
	}{}
	pipeline.ParseArgumentsIncludingParents(&arguments, "ssh", run)

	if arguments.Host != nil {
		run.Log.Debug(
			fields.Symbol("👨‍💻"),
			fields.Message(arguments.Host),
			fields.Middleware(sshMiddleware),
		)
		prefixWithSSHHost(run, *arguments.Host)
	}

	next(run)
}

func prefixWithSSHHost(run *pipeline.Run, sshHost string) {
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
