package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
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
	mutex   *sync.RWMutex
}

func newDefaultCommandExecutor() *defaultCommandExecutor {
	return &defaultCommandExecutor{
		env:     os.Environ(),
		stopped: false,
		mutex:   &sync.RWMutex{},
	}
}

func (executor *defaultCommandExecutor) Init(name string, arg ...string) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	executor.command = exec.Command(name, arg...)
	executor.command.Env = executor.env
}

func (executor *defaultCommandExecutor) Start() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return fmt.Errorf("cannot start cleared command")
	}
	return executor.command.Start()
}

func (executor *defaultCommandExecutor) CmdStdin() io.WriteCloser {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return nil
	}
	stdin, _ := executor.command.StdinPipe()
	return stdin
}

func (executor *defaultCommandExecutor) CmdStdout() io.Reader {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return nil
	}
	stdout, _ := executor.command.StdoutPipe()
	return stdout
}

func (executor *defaultCommandExecutor) CmdStderr() io.Reader {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return nil
	}
	stderr, _ := executor.command.StderrPipe()
	return stderr
}

func (executor *defaultCommandExecutor) Wait() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return fmt.Errorf("cannot wait for cleared command")
	}
	return executor.command.Wait()
}

func (executor *defaultCommandExecutor) Kill() error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil || executor.command.Process == nil {
		return nil
	}
	result := executor.command.Process.Kill()
	executor.command = nil
	return result
}

func (executor *defaultCommandExecutor) Clear() {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	executor.command = nil
}

func (executor *defaultCommandExecutor) String() string {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if executor.command == nil {
		return ""
	}
	return executor.command.String()
}
