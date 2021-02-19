package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type commandExecutor interface {
	Init(name string, arg ...string)
	Kill() error
	Clear()
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
	if executor.command == nil {
		return fmt.Errorf("cannot start cleared command")
	}
	return executor.command.Start()
}

func (executor *defaultCommandExecutor) CmdStdin() io.WriteCloser {
	if executor.command == nil {
		return nil
	}
	stdin, _ := executor.command.StdinPipe()
	return stdin
}

func (executor *defaultCommandExecutor) CmdStdout() io.Reader {
	if executor.command == nil {
		return nil
	}
	stdout, _ := executor.command.StdoutPipe()
	return stdout
}

func (executor *defaultCommandExecutor) CmdStderr() io.Reader {
	if executor.command == nil {
		return nil
	}
	stderr, _ := executor.command.StderrPipe()
	return stderr
}

func (executor *defaultCommandExecutor) Wait() error {
	if executor.command == nil {
		return fmt.Errorf("cannot wait for cleared command")
	}
	return executor.command.Wait()
}

func (executor *defaultCommandExecutor) Kill() error {
	if executor.command == nil || executor.command.Process == nil {
		return nil
	}
	result := executor.command.Process.Kill()
	executor.command = nil
	return result
}

func (executor *defaultCommandExecutor) Clear() {
	executor.command = nil
}

func (executor *defaultCommandExecutor) String() string {
	if executor.command == nil {
		return ""
	}
	return executor.command.String()
}
