package shell

import (
	"bufio"
	"bytes"
	"fmt"
	customio "github.com/Layer9Berlin/pipedream/src/custom/io"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

func TestShell_NonRunnable(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.NotContains(t, run.Log.String(), "shell")
}

func TestShell_ChangeDir(t *testing.T) {
	identifier := "command-identifier"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"shell": map[string]interface{}{
			"dir": "test",
			"run": "something",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	executor, shellMiddleware := NewTestShellMiddleware()
	shellMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "something")
	require.Equal(t, "sh", executor.StartCommand)
	require.Equal(t, []string{"-c", "cd test && something"}, executor.StartArgs)
	require.Contains(t, logString, "shell")
	require.Contains(t, logString, "cd test")
	require.Contains(t, logString, "Command Identifier")
}

func TestShell_RunWithArguments(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"args": []interface{}{
				map[string]interface{}{
					"test": "value",
				},
				"another_value",
				map[string]interface{}{
					"test2": "value2",
					"test3": "value3",
					"a":     true,
					"b":     false,
					"c":     "value",
				},
				"--long=raw",
				"--long raw-with-space",
				"-sraw",
				"-s=raw",
				map[string]interface{}{
					"--long-from-map=": "value",
					"-r":               true,
					"-s":               "without-space",
					"s":                "with space",
					"raw":              "test",
					"test1":            true,
					"z":                false,
				},
			},
			"run": "something",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.TraceLevel)
	executor, shellMiddleware := NewTestShellMiddleware()
	shellMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, "sh", executor.StartCommand)
	require.Equal(t, []string{"-c", "something --test=\"value\" another_value --test2=\"value2\" --test3=\"value3\" -a -c \"value\" --long=raw --long raw-with-space -sraw -s=raw --long-from-map=\"value\" --raw=\"test\" --test1 -r -s \"with space\" -s\"without-space\""}, executor.StartArgs)
}

func TestShell_InvalidArguments(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"args": []interface{}{
				[]interface{}{
					"invalid",
				},
			},
			"run": "something",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	_, shellMiddleware := NewTestShellMiddleware()
	shellMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			require.Fail(t, "not expected to be called")
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "anonymous")
	require.Contains(t, logString, "shell")
}

func TestShell_Login(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"login": true,
			"run":   "something",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.TraceLevel)
	executor, shellMiddleware := NewTestShellMiddleware()
	shellMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, []string{"-l", "-c", "something"}, executor.StartArgs)
	require.Contains(t, run.Log.String(), "shell")
}

func TestShell_NonZeroExitCode(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"login": true,
			"run":   "something",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.TraceLevel)
	executor, shellMiddleware := NewTestShellMiddleware()
	executor.WaitError = &exec.ExitError{}
	shellMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	// non-zero exit codes are now reported as warnings, not errors
	// (until a more sophisticated implementation of the `catch` middleware
	// allows us to actually deal with and clear exit codes)
	require.Equal(t, 1, run.Log.WarnCount())
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, []string{"-l", "-c", "something"}, executor.StartArgs)
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, -1, *run.ExitCode)
}

func TestShell_Interactive_userInput(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"interactive": true,
			"run":         "something",
		},
	}, nil, nil)

	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	executor := NewTestCommandExecutor()

	osStdin := customio.NewSynchronizedBuffer()
	osStdout := customio.NewSynchronizedBuffer()
	osStderr := customio.NewSynchronizedBuffer()
	shellMiddleware := Middleware{
		osStdin:         osStdin,
		osStdout:        osStdout,
		osStderr:        osStderr,
		ExecutorCreator: func() commandExecutor { return executor },
	}
	run.Stdout.Replace(strings.NewReader("Please confirm (Y/n):\n"))
	run.Stderr.Replace(strings.NewReader("Middleware-defined stderr\n"))
	// we simulate the following sequence of events:
	// 1) The executing command shows a prompt to the user in its output, which is mirrored in the OS stdout.
	// 2) The user enters a string, which is passed to the command through the OS stdin.
	// 3) The command's input is set to the string entered by the user, its output is set to the prompt.
	// 4) In addition, the command may provide stderr output.
	shellMiddleware.Apply(
		run,
		func(nextRun *pipeline.Run) {
			nextExecuted = true
		},
		nil,
	)
	go func() {
		// simulate user input to OS stdin
		_, err := io.WriteString(osStdin, "y\n")
		require.Nil(t, err)
	}()
	executor.WaitGroup.Add(1)
	go func() {
		defer executor.WaitGroup.Done()
		// simulate shell command consuming its input
		scanner := bufio.NewScanner(executor.StdinReader)
		for scanner.Scan() {
			// the shell command's input is the pipe's input plus the OS stdin
			require.Equal(t, "y", scanner.Text())
			// now pretend the command has finished
			// it will indicate this through closing CmdStdin()
			_ = executor.CmdStdin().Close()
			break
		}
	}()
	run.Start()
	run.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	//the pipe's output
	require.Equal(t, "Please confirm (Y/n):\n", run.Stdout.String())
	//the (simulated) OS stdout output
	require.Equal(t, "Please confirm (Y/n):\n", osStdout.String())
	//the pipe's stderr
	require.Equal(t, "Middleware-defined stderr\n", run.Stderr.String())
	//the (simulated) OS stderr output
	require.Equal(t, "Middleware-defined stderr\n", osStderr.String())
}

func TestShell_Interactive_userInputError(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"interactive": true,
			"run":         "something",
		},
	}, nil, nil)

	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	executor := NewTestCommandExecutor()
	osStdin := customio.NewErrorReader()
	osStdout := new(bytes.Buffer)
	osStderr := new(bytes.Buffer)
	shellMiddleware := Middleware{
		osStdin:         osStdin,
		osStdout:        osStdout,
		osStderr:        osStderr,
		ExecutorCreator: func() commandExecutor { return executor },
	}
	run.Stdout.Replace(strings.NewReader("Please confirm (Y/n):\n"))
	run.Stderr.Replace(strings.NewReader("Middleware-defined stderr\n"))
	// we simulate the following sequence of events:
	// 1) The executing command shows a prompt to the user in its output, which is mirrored in the OS stdout.
	// 2) The user enters a string, which is passed to the command through the OS stdin.
	// 3) The command's input is set to the string entered by the user, its output is set to the prompt.
	// 4) In addition, the command may provide stderr output.
	shellMiddleware.Apply(
		run,
		func(nextRun *pipeline.Run) {
			nextExecuted = true
		},
		nil,
	)
	executor.WaitGroup.Add(1)
	go func() {
		defer executor.WaitGroup.Done()
		_ = executor.CmdStdin().Close()
	}()
	run.Start()
	run.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	//the pipe's output
	require.Equal(t, "Please confirm (Y/n):\n", run.Stdout.String())
	//the (simulated) OS stdout output
	require.Equal(t, "Please confirm (Y/n):\n", osStdout.String())
	//the pipe's stderr
	require.Equal(t, "Middleware-defined stderr\n", run.Stderr.String())
	//the (simulated) OS stderr output
	require.Equal(t, "Middleware-defined stderr\n", osStderr.String())
}

func TestShell_Interactive_pipeInput(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"interactive": true,
			"run":         "something",
		},
	}, nil, nil)

	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	executor := NewTestCommandExecutor()
	osStdin := new(bytes.Buffer)
	osStdout := new(bytes.Buffer)
	osStderr := new(bytes.Buffer)
	shellMiddleware := Middleware{
		osStdin:         osStdin,
		osStdout:        osStdout,
		osStderr:        osStderr,
		ExecutorCreator: func() commandExecutor { return executor },
	}
	run.Stdin.Replace(strings.NewReader("y\n"))
	run.Stdout.Replace(strings.NewReader("Please confirm (Y/n):\n"))
	run.Stderr.Replace(strings.NewReader("Middleware-defined stderr\n"))
	// we simulate the following sequence of events:
	// 1) The executing command shows a prompt to the user in its output, which is mirrored in the OS stdout.
	// 2) A string is passed to the pipe through its input connection (not the OS stdin).
	// 3) The command's input is set to the string entered by the user, its output is set to the prompt.
	// 4) In addition, the command may provide stderr output.
	shellMiddleware.Apply(
		run,
		func(nextRun *pipeline.Run) {
			nextExecuted = true
		},
		nil,
	)
	executor.WaitGroup.Add(1)
	go func() {
		defer executor.WaitGroup.Done()
		// simulate shell command consuming its input
		scanner := bufio.NewScanner(executor.StdinReader)
		for scanner.Scan() {
			// the shell command's input is the pipe's input plus the OS stdin
			require.Equal(t, "y", scanner.Text())
			// now pretend the command has finished
			// it will indicate this through closing CmdStdin()
			_ = executor.CmdStdin().Close()
		}
	}()
	run.Start()
	run.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	//the pipe's output
	require.Equal(t, "Please confirm (Y/n):\n", run.Stdout.String())
	//the (simulated) OS stdout output
	require.Equal(t, "Please confirm (Y/n):\n", osStdout.String())
	//the pipe's stderr
	require.Equal(t, "Middleware-defined stderr\n", run.Stderr.String())
	//the (simulated) OS stderr output
	require.Equal(t, "Middleware-defined stderr\n", osStderr.String())
}

func TestShell_Interactive_pipeInputError(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"interactive": true,
			"run":         "something",
		},
	}, nil, nil)

	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	executor := NewTestCommandExecutor()
	osStdin := new(bytes.Buffer)
	osStdout := new(bytes.Buffer)
	osStderr := new(bytes.Buffer)
	shellMiddleware := Middleware{
		osStdin:         osStdin,
		osStdout:        osStdout,
		osStderr:        osStderr,
		ExecutorCreator: func() commandExecutor { return executor },
	}
	run.Stdin.Replace(customio.NewErrorReader())
	run.Stdout.Replace(strings.NewReader("Please confirm (Y/n):\n"))
	run.Stderr.Replace(strings.NewReader("Middleware-defined stderr\n"))
	// we simulate the following sequence of events:
	// 1) The executing command shows a prompt to the user in its output, which is mirrored in the OS stdout.
	// 2) A string is passed to the pipe through its input connection (not the OS stdin).
	// 3) The command's input is set to the string entered by the user, its output is set to the prompt.
	// 4) In addition, the command may provide stderr output.
	shellMiddleware.Apply(
		run,
		func(nextRun *pipeline.Run) {
			nextExecuted = true
		},
		nil,
	)
	executor.WaitGroup.Add(1)
	go func() {
		defer executor.WaitGroup.Done()
		_ = executor.CmdStdin().Close()
	}()
	run.Start()
	run.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Equal(t, "test error", run.Log.LastError().Error())
	require.Contains(t, run.Log.String(), "shell")
	//the pipe's output
	require.Equal(t, "Please confirm (Y/n):\n", run.Stdout.String())
	//the (simulated) OS stdout output
	require.Equal(t, "Please confirm (Y/n):\n", osStdout.String())
	//the pipe's stderr
	require.Equal(t, "Middleware-defined stderr\n", run.Stderr.String())
	//the (simulated) OS stderr output
	require.Equal(t, "Middleware-defined stderr\n", osStderr.String())
}

func TestShell_WaitError(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "something",
		},
	}, nil, nil)

	executor := NewTestCommandExecutor()
	executor.WaitError = fmt.Errorf("test error")
	shellMiddleware := Middleware{
		osStdin:         new(bytes.Buffer),
		osStdout:        new(bytes.Buffer),
		osStderr:        new(bytes.Buffer),
		ExecutorCreator: func() commandExecutor { return executor },
	}
	shellMiddleware.Apply(
		run,
		func(nextRun *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "test error")
}

func TestShell_UnmockedCommand(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "echo \"Test\"",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Start()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, "Test\n", run.Stdout.String())
}

func TestShell_Cancel(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "read",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	err := run.Cancel()
	require.Nil(t, err)
	run.Wait()

	require.Contains(t, run.Log.String(), "shell")
}

func TestShell_PrintfRun(t *testing.T) {
	run, err := pipeline.NewRun(
		nil,
		map[string]interface{}{
			"shell": map[string]interface{}{
				"run": "printf \"test\"",
			},
		},
		nil,
		nil,
	)
	require.Nil(t, err)

	run.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, "test", run.Stdout.String())
}

type TestCommandExecutor struct {
	StartCommand      string
	StartArgs         []string
	StdinReader       io.Reader
	StdinWriteCloser  io.WriteCloser
	StdoutReader      io.Reader
	StdoutWriteCloser io.WriteCloser
	StderrReader      io.Reader
	StderrWriteCloser io.WriteCloser
	WaitGroup         *sync.WaitGroup
	WaitError         error
	KillError         error
	Mutex             *sync.RWMutex
}

func (executor *TestCommandExecutor) Init(name string, arg ...string) {
	executor.Mutex.Lock()
	defer executor.Mutex.Unlock()
	executor.StartCommand = name
	executor.StartArgs = arg
}

func (executor *TestCommandExecutor) Start() error {
	go func() {
		executor.WaitGroup.Wait()
		executor.Mutex.RLock()
		defer executor.Mutex.RUnlock()
		_ = executor.StdinWriteCloser.Close()
		_ = executor.StdoutWriteCloser.Close()
		_ = executor.StderrWriteCloser.Close()
	}()
	return nil
}

func (executor *TestCommandExecutor) CmdStdin() io.WriteCloser {
	executor.Mutex.RLock()
	defer executor.Mutex.RUnlock()
	return executor.StdinWriteCloser
}

func (executor *TestCommandExecutor) CmdStdout() io.Reader {
	executor.Mutex.RLock()
	defer executor.Mutex.RUnlock()
	return executor.StdoutReader
}

func (executor *TestCommandExecutor) CmdStderr() io.Reader {
	executor.Mutex.RLock()
	defer executor.Mutex.RUnlock()
	return executor.StderrReader
}

func (executor *TestCommandExecutor) Wait() error {
	executor.Mutex.Lock()
	defer executor.Mutex.Unlock()
	executor.WaitGroup.Wait()
	return executor.WaitError
}

func (executor *TestCommandExecutor) Kill() error {
	executor.Mutex.RLock()
	defer executor.Mutex.RUnlock()
	return executor.KillError
}

func (executor *TestCommandExecutor) String() string {
	executor.Mutex.RLock()
	defer executor.Mutex.RUnlock()
	return fmt.Sprintf("%v %v", executor.StartCommand, executor.StartArgs)
}

func (executor *TestCommandExecutor) Clear() {
}

func NewTestCommandExecutor() *TestCommandExecutor {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	return &TestCommandExecutor{
		StdinReader:       bufio.NewReader(stdinReader),
		StdinWriteCloser:  stdinWriter,
		StdoutReader:      bufio.NewReader(stdoutReader),
		StdoutWriteCloser: stdoutWriter,
		StderrReader:      bufio.NewReader(stderrReader),
		StderrWriteCloser: stderrWriter,
		WaitGroup:         &sync.WaitGroup{},
		Mutex:             &sync.RWMutex{},
	}
}

func NewTestShellMiddleware() (*TestCommandExecutor, Middleware) {
	executor := NewTestCommandExecutor()
	return executor, NewMiddlewareWithExecutorCreator(func() commandExecutor { return executor })
}
