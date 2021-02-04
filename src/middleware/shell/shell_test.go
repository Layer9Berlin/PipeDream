package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"os/exec"
	"sync"
	"testing"
)

func TestShell_NonRunnable(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.DebugLevel)
	NewShellMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.NotContains(t, run.Log.String(), "shell")
}

func TestShell_ChangeDir(t *testing.T) {
	identifier := "command-identifier"
	run, _ := models.NewPipelineRun(&identifier, map[string]interface{}{
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
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "something")
	require.Equal(t, "sh", executor.StartCommand)
	require.Equal(t, []string{"-c", "cd test && something"}, executor.StartArgs)
	require.Contains(t, run.Log.String(), "shell")
	require.Contains(t, run.Log.String(), "cd test")
	require.Contains(t, run.Log.String(), "Command Identifier")
}

func TestShell_RunWithArguments(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
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
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, "sh", executor.StartCommand)
	require.Equal(t, []string{"-c", "something --test=\"value\" another_value --test2=\"value2\" --test3=\"value3\" -a -c \"value\" --long=raw --long raw-with-space -sraw -s=raw --long-from-map=\"value\" --raw=\"test\" --test1 -r -s \"with space\" -s\"without-space\""}, executor.StartArgs)
}

func TestShell_InvalidArguments(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
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
		func(run *models.PipelineRun) {
			require.Fail(t, "not expected to be called")
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "anonymous")
	require.Contains(t, run.Log.String(), "shell")
}

func TestShell_Login(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
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
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, []string{"-l", "-c", "something"}, executor.StartArgs)
	require.Contains(t, run.Log.String(), "shell")
}

func TestShell_NonZeroExitCode(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
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
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "command exited with non-zero exit code")
	require.Equal(t, []string{"-l", "-c", "something"}, executor.StartArgs)
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, -1, *run.ExitCode)
}

//func TestShell_Interactive(t *testing.T) {
//	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
//		"shell": map[string]interface{}{
//			"interactive": true,
//			"run":         "something",
//		},
//	}, nil, nil)
//
//	nextExecuted := false
//	run.Log.SetLevel(logrus.DebugLevel)
//	executor := NewTestCommandExecutor()
//	osStdin := new(bytes.Buffer)
//	osStdout := new(bytes.Buffer)
//	osStderr := new(bytes.Buffer)
//	shellMiddleware := ShellMiddleware{
//		osStdin:         osStdin,
//		osStdout:        osStdout,
//		osStderr:        osStderr,
//		ExecutorCreator: func() CommandExecutor { return executor },
//	}
//	shellCommandInput := ""
//	go func() {
//		// simulate user input to OS stdin
//		_, err := io.WriteString(osStdin, "y\n")
//		require.Nil(t, err)
//	}()
//	executor.WaitGroup.Add(1)
//	go func() {
//		// simulate shell command consuming its input and finishing after two lines of input
//		reader := bufio.NewReader(executor.StdinReader)
//		// the first line is available immediately
//		chunk, err := reader.ReadString('\n')
//		require.Nil(t, err)
//		shellCommandInput += chunk
//		for {
//			// the second may be entered by the simulated user at any time
//			chunk, err := reader.ReadString('\n')
//			shellCommandInput += chunk
//			if err != io.EOF {
//				break
//			}
//			time.Sleep(100 * time.Millisecond)
//		}
//		require.Nil(t, err)
//		executor.WaitGroup.Done()
//	}()
//	executor.WaitGroup.Add(1)
//	go func() {
//		defer executor.WaitGroup.Done()
//		// simulate shell command giving some output
//		_, _ = io.WriteString(executor.StdoutWriteCloser, "(Additional prompt):\n")
//	}()
//	executor.WaitGroup.Add(1)
//	go func() {
//		defer executor.WaitGroup.Done()
//		// simulate shell command giving some stderr output
//		_, _ = io.WriteString(executor.StderrWriteCloser, "(Additional stderr):\n")
//	}()
//	shellMiddleware.Apply(
//		run,
//		func(nextRun *models.PipelineRun) {
//			// the input might come from a previous pipe
//			nextRun.Stdin.Replace(strings.NewReader("Please confirm (Y/n):\n"))
//			//this could be set by other middleware
//			nextRun.Stdout.Replace(strings.NewReader("Middleware-defined output to be shown to user\n"))
//			nextRun.Stderr.Replace(strings.NewReader("Middleware-defined stderr to be shown to user\n"))
//			nextExecuted = true
//		},
//		nil,
//	)
//	run.Close()
//	run.Wait()
//
//	require.True(t, nextExecuted)
//	require.Equal(t, 0, run.Log.ErrorCount())
//	require.Contains(t, run.Log.String(), "shell (Shell Command Runner)")
//	// the pipe's input
//	require.Equal(t, "y\n", run.Stdin.String())
//	// the shell command's input is the pipe's input plus the OS stdin
//	require.Equal(t, "Please confirm (Y/n):\ny\n", shellCommandInput)
//	//the pipe's output
//	require.Equal(t, "Middleware-defined output to be shown to user\n(Additional prompt):\n", run.Stdout.String())
//	//the (simulated) OS stdout output
//	require.Equal(t, "Middleware-defined output to be shown to user\n(Additional prompt):\n", osStdout.String())
//	//the pipe's stderr
//	require.Equal(t, "Middleware-defined stderr to be shown to user\n(Additional stderr):\n", run.Stderr.String())
//	//the (simulated) OS stderr output
//	require.Equal(t, "Middleware-defined stderr to be shown to user\n(Additional stderr):\n", osStderr.String())
//}

func TestShell_WaitError(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "something",
		},
	}, nil, nil)

	executor := NewTestCommandExecutor()
	executor.WaitError = fmt.Errorf("test error")
	shellMiddleware := ShellMiddleware{
		osStdin:         new(bytes.Buffer),
		osStdout:        new(bytes.Buffer),
		osStderr:        new(bytes.Buffer),
		ExecutorCreator: func() CommandExecutor { return executor },
	}
	shellMiddleware.Apply(
		run,
		func(nextRun *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "test error")
}

func TestShell_UnmockedCommand(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "echo \"Test\"",
		},
	}, nil, nil)

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	nextExecuted := false
	run.Log.SetLevel(logrus.TraceLevel)
	NewShellMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {
			defer testWaitGroup.Done()
			nextExecuted = true
		},
		nil,
	)
	run.Close()
	run.Wait()
	testWaitGroup.Wait()

	require.True(t, nextExecuted)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "shell")
	require.Equal(t, "Test\n", run.Stdout.String())
}

func TestShell_CancelHook(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "read",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewShellMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {
		},
		nil,
	)
	run.Close()
	err := run.Cancel()
	require.Nil(t, err)
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "command exited with non-zero exit code")
	require.Contains(t, run.Log.String(), "shell")
}

func TestShell_PrintfRun(t *testing.T) {
	run, err := models.NewPipelineRun(
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
	NewShellMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {
		},
		nil,
	)
	run.Close()
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
}

func (executor *TestCommandExecutor) Init(name string, arg ...string) {
	executor.StartCommand = name
	executor.StartArgs = arg
}

func (executor *TestCommandExecutor) Start() error {
	go func() {
		executor.WaitGroup.Wait()
		_ = executor.StdinWriteCloser.Close()
		_ = executor.StdoutWriteCloser.Close()
		_ = executor.StderrWriteCloser.Close()
	}()
	return nil
}

func (executor *TestCommandExecutor) CmdStdin() io.WriteCloser {
	return executor.StdinWriteCloser
}

func (executor *TestCommandExecutor) CmdStdout() io.Reader {
	return executor.StdoutReader
}

func (executor *TestCommandExecutor) CmdStderr() io.Reader {
	return executor.StderrReader
}

func (executor *TestCommandExecutor) Wait() error {
	executor.WaitGroup.Wait()
	return executor.WaitError
}

func (executor *TestCommandExecutor) Kill() error {
	return executor.KillError
}

func (executor *TestCommandExecutor) String() string {
	return fmt.Sprintf("%v %v", executor.StartCommand, executor.StartArgs)
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
	}
}

func NewTestShellMiddleware() (*TestCommandExecutor, ShellMiddleware) {
	executor := NewTestCommandExecutor()
	return executor, NewShellMiddlewareWithExecutorCreator(func() CommandExecutor { return executor })
}
