package shell

import (
	"io"
	"os"
	"os/exec"
)

type commandExecutor interface {
	Init(name string, arg ...string)
	Kill() error
	CmdStdin() io.WriteCloser
	CmdStdout() io.Reader
	CmdStderr() io.Reader
	Start() error
	String() string
	Wait() error
}

type defaultCommandExecutor struct {
	command *exec.Cmd
	env     []string
	stopped bool
}

func newDefaultCommandExecutor() *defaultCommandExecutor {
	return &defaultCommandExecutor{
		env:     os.Environ(),
		stopped: false,
	}
}

func (executor *defaultCommandExecutor) Init(name string, arg ...string) {
	executor.command = exec.Command(name, arg...)
	executor.command.Env = executor.env
}

func (executor *defaultCommandExecutor) Start() error {
	return executor.command.Start()
}

func (executor *defaultCommandExecutor) CmdStdin() io.WriteCloser {
	stdin, _ := executor.command.StdinPipe()
	return stdin
}

func (executor *defaultCommandExecutor) CmdStdout() io.Reader {
	stdout, _ := executor.command.StdoutPipe()
	return stdout
}

func (executor *defaultCommandExecutor) CmdStderr() io.Reader {
	stderr, _ := executor.command.StderrPipe()
	return stderr
}

func (executor *defaultCommandExecutor) Wait() error {
	return executor.command.Wait()
}

func (executor *defaultCommandExecutor) Kill() error {
	if executor.command.Process == nil {
		return nil
	}
	return executor.command.Process.Kill()
}

func (executor *defaultCommandExecutor) String() string {
	return executor.command.String()
}
