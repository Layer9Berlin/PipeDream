package middleware

import (
	"bytes"
	"fmt"
	customio "github.com/Layer9Berlin/pipedream/src/custom/io"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/parsing"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestExecutionContext_CancelAll(t *testing.T) {
	executionContext := NewExecutionContext()
	run1, _ := pipeline.NewRun(nil, nil, nil, nil)
	run2, _ := pipeline.NewRun(nil, nil, nil, nil)
	run3, _ := pipeline.NewRun(nil, nil, nil, nil)
	run1.DontCompleteBefore(func() {
		time.Sleep(1 * time.Second)
	})
	run1.Start()
	run2.Start()
	executionContext.runs = []*pipeline.Run{run1, run2, run3}
	err := executionContext.CancelAll()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "cancelling a run that has not yet started")
	require.True(t, executionContext.runs[0].Cancelled())
	require.True(t, executionContext.runs[1].Cancelled())
	require.True(t, executionContext.runs[2].Cancelled())
}

func TestExecutionContext_FullRun_WithoutOptions(t *testing.T) {
	executionContext := NewExecutionContext()
	run := executionContext.FullRun()
	require.NotNil(t, run)
	require.Equal(t, run, executionContext.rootRun)
}

func TestExecutionContext_FullRun_WithDefinitionsLookupOption(t *testing.T) {
	arguments := map[string]interface{}{
		"key": "value",
	}
	executionContext := NewExecutionContext(WithDefinitionsLookup(map[string][]pipeline.Definition{
		"test": {
			{
				DefinitionArguments: arguments,
			},
		},
	}))
	identifier := "test"
	run := executionContext.FullRun(WithIdentifier(&identifier))
	require.NotNil(t, run)
	require.Equal(t, arguments, run.ArgumentsCopy())
}

func TestExecutionContext_FullRun_WithUnmergeableArguments(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("failed to encounter expected panic")
		}
	}()
	arguments1 := map[string]interface{}{
		"key": []interface{}{},
	}
	arguments2 := map[string]interface{}{
		"key": map[string]interface{}{},
	}
	executionContext := ExecutionContext{
		Definitions: map[string][]pipeline.Definition{
			"test": {
				{
					DefinitionArguments: arguments1,
				},
			},
		},
	}
	identifier := "test"
	run := executionContext.FullRun(WithIdentifier(&identifier), WithArguments(arguments2))
	require.Nil(t, run)
}

func TestExecutionContext_FullRun_WithSetupFunction(t *testing.T) {
	setupCalled := false
	setupFunc := func(run *pipeline.Run) {
		setupCalled = true
	}
	executionContext := NewExecutionContext()
	identifier := "test"
	run := executionContext.FullRun(
		WithIdentifier(&identifier),
		WithSetupFunc(setupFunc),
	)
	require.NotNil(t, run)
	require.True(t, setupCalled)
}

func TestExecutionContext_FullRun_WithTearDownFunction(t *testing.T) {
	tearDownCalled := false
	tearDownFunc := func(run *pipeline.Run) {
		tearDownCalled = true
	}
	executionContext := NewExecutionContext()
	identifier := "test"
	run := executionContext.FullRun(
		WithIdentifier(&identifier),
		WithTearDownFunc(tearDownFunc),
	)
	require.NotNil(t, run)
	require.True(t, tearDownCalled)
}

func TestExecutionContext_FullRun_WithParentRun(t *testing.T) {
	parentRun, _ := pipeline.NewRun(nil, nil, nil, nil)
	executionContext := NewExecutionContext()
	identifier := "test"
	run := executionContext.FullRun(
		WithIdentifier(&identifier),
		WithParentRun(parentRun),
	)
	require.NotNil(t, run)
	require.Equal(t, parentRun, run.Parent)
}

func TestExecutionContext_FullRun_WithLogWriter(t *testing.T) {
	logReader, logWriter := io.Pipe()
	executionContext := NewExecutionContext(WithExecutionFunction(func(run *pipeline.Run) {
		run.Log.Info(fields.Message("test"))
	}),
	)
	identifier := "test"
	run := executionContext.FullRun(
		WithIdentifier(&identifier),
		WithLogWriter(logWriter),
		WithSetupFunc(func(run *pipeline.Run) {
			run.Log.SetLevel(logrus.InfoLevel)
		}),
	)
	log := ""
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		logData, _ := ioutil.ReadAll(logReader)
		log = string(logData)
	}()
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.NotNil(t, run)
	require.Contains(t, log, "test")
}

func TestExecutionContext_UnwindStack(t *testing.T) {
	middleware1 := NewFakeMiddleware()
	middleware2 := NewFakeMiddleware()
	stack := []Middleware{
		middleware1,
		middleware2,
	}

	executionContext := ExecutionContext{
		MiddlewareStack: stack,
	}
	executionContext.unwindStack(nil, 0)
	require.Equal(t, 1, middleware1.CallCount)
	require.Equal(t, 1, middleware2.CallCount)
}

func TestExecutionContext_PipelineFileAtPath(t *testing.T) {
	executionContext := ExecutionContext{
		PipelineFiles: []pipeline.File{
			{
				Path: "test1",
			},
			{
				Path: "test2",
			},
			{
				Path: "test3",
			},
		},
	}
	file, err := executionContext.PipelineFileAtPath("test2")
	require.Nil(t, err)
	require.NotNil(t, file)
}

func TestExecutionContext_PipelineFileAtPath_NotFound(t *testing.T) {
	executionContext := ExecutionContext{
		PipelineFiles: []pipeline.File{
			{
				FileName: "test1",
			},
			{
				FileName: "test2",
			},
			{
				FileName: "test3",
			},
		},
	}
	file, err := executionContext.PipelineFileAtPath("test4")
	require.NotNil(t, err)
	require.Nil(t, file)
}

func TestExecutionContext_LookUpPipelineDefinition(t *testing.T) {
	definitionsLookup := map[string][]pipeline.Definition{
		"test1": {
			{
				FileName: "test1.file",
				Public:   false,
			},
			{
				FileName: "test2.file",
				Public:   true,
			},
			{
				FileName: "test3.file",
				Public:   false,
			},
		},
		"test2": {},
	}
	definition, found := LookUpPipelineDefinition(definitionsLookup, "test1", "test3.file")
	require.True(t, found)
	require.Equal(t, "test3.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test1", "test4.file")
	require.True(t, found)
	require.Equal(t, "test2.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test1", "test1.file")
	require.True(t, found)
	require.Equal(t, "test1.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test2", "test1.file")
	require.False(t, found)
	require.Nil(t, definition)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "invalid", "test1.file")
	require.False(t, found)
	require.Nil(t, definition)
}

func TestExecutionContext_Execute(t *testing.T) {
	buffer := new(bytes.Buffer)
	executionContext := NewExecutionContext()
	executionContext.Execute("test", buffer, new(bytes.Buffer))
	require.NotContains(t, buffer.String(), "====== LOGS ======")
	require.Contains(t, buffer.String(), "===== RESULT =====")
}

func TestExecutionContext_SetUpPipelines(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(
					func(pattern string) ([]string, error) {
						return []string{
							"test1.pipe",
							"test2.pipe",
						}, nil
					}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte(""), nil
				}))))
	buffer := new(bytes.Buffer)
	executionContext.Log.SetOutput(buffer)

	err := executionContext.SetUpPipelines("")
	require.Nil(t, err)
	require.NotNil(t, executionContext.PipelineFiles)

	require.Equal(t, "", buffer.String())
}

func TestExecutionContext_SetUpPipelines_BuiltInPipelineFilePathsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{}, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines("")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_ParseBuiltInPipelineFilesError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return nil, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines("")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_UserPipelineFilePathsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					if strings.Contains(pattern, "pipes/**") {
						return []string{"test.file"}, nil
					}
					return []string{}, fmt.Errorf("test error")
				}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte{}, nil
				}),
			)))
	err := executionContext.SetUpPipelines("")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_RecursivelyAddImportsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte{}, nil
				}),
				parsing.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return nil, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines("")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_ParsePipelineFilesError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test1"}, nil
				}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					if filename == "test.file" {
						return nil, fmt.Errorf("test error")
					}
					return []byte{}, nil
				}),
				parsing.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
			)))
	err := executionContext.SetUpPipelines("")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_LogFullRun(t *testing.T) {
	previousLogLevel := logging.UserPipeLogLevel
	logging.UserPipeLogLevel = logrus.TraceLevel
	defer func() {
		logging.UserPipeLogLevel = previousLogLevel
	}()
	pipedWriteCloser := customio.NewPipedWriteCloser()
	executionContext := NewExecutionContext()
	run := executionContext.FullRun(WithLogWriter(pipedWriteCloser))

	run.Wait()
	pipedWriteCloser.Wait()
	require.Contains(t, pipedWriteCloser.String(), "full run")
}

func TestExecutionContext_FullRun_LogStderr(t *testing.T) {
	previousLogLevel := logging.UserPipeLogLevel
	logging.UserPipeLogLevel = logrus.TraceLevel
	defer func() {
		logging.UserPipeLogLevel = previousLogLevel
	}()
	pipedWriteCloser := customio.NewPipedWriteCloser()
	executionContext := NewExecutionContext()
	run := executionContext.FullRun(
		WithLogWriter(pipedWriteCloser),
		WithTearDownFunc(func(run *pipeline.Run) {
			run.Stderr.Replace(strings.NewReader("test output"))
		}),
	)
	run.Wait()
	pipedWriteCloser.Wait()
	require.Contains(t, pipedWriteCloser.String(), "test output")
}

func TestExecutionContext_FullRun_LogStdin(t *testing.T) {
	previousLogLevel := logging.UserPipeLogLevel
	logging.UserPipeLogLevel = logrus.TraceLevel
	defer func() {
		logging.UserPipeLogLevel = previousLogLevel
	}()
	pipedWriteCloser := customio.NewPipedWriteCloser()
	executionContext := NewExecutionContext()
	run := executionContext.FullRun(
		WithLogWriter(pipedWriteCloser),
		WithSetupFunc(func(run *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("test input"))
		}),
	)
	run.Wait()
	pipedWriteCloser.Wait()
	require.Contains(t, pipedWriteCloser.String(), "test input")
}

func TestExecutionContext_FullRun_LogStdout(t *testing.T) {
	previousLogLevel := logging.UserPipeLogLevel
	logging.UserPipeLogLevel = logrus.TraceLevel
	defer func() {
		logging.UserPipeLogLevel = previousLogLevel
	}()
	pipedWriteCloser := customio.NewPipedWriteCloser()
	executionContext := NewExecutionContext()
	run := executionContext.FullRun(
		WithLogWriter(pipedWriteCloser),
		WithTearDownFunc(func(run *pipeline.Run) {
			run.Stdout.Replace(strings.NewReader("test output"))
		}),
	)
	run.Wait()
	pipedWriteCloser.Wait()
	require.Contains(t, pipedWriteCloser.String(), "test output")
}

func TestExecutionContext_defaultUserPrompt(t *testing.T) {
	pipedWriteCloser := customio.NewPipedWriteCloser()
	resultIndex, resultString, err := defaultUserPrompt("test", []string{"option1", "option2"}, 1, 5, ioutil.NopCloser(strings.NewReader("\n")), pipedWriteCloser)
	require.Nil(t, err)
	require.Equal(t, 1, resultIndex)
	require.Equal(t, "option2", resultString)
}

func TestExecutionContext_CancelError(t *testing.T) {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	executionContext := NewExecutionContext()
	run, _ := pipeline.NewRun(nil, nil, nil, nil)
	run.AddCancelHook(func() error {
		return fmt.Errorf("test error")
	})
	executionContext.runs = []*pipeline.Run{
		run,
	}
	stdoutWriter := customio.NewPipedWriteCloser()
	stderrWriter := customio.NewPipedWriteCloser()
	executionContext.SetUpCancelHandler(stdoutWriter, stderrWriter, func() {
		waitGroup.Done()
	})
	executionContext.interruptChannel <- syscall.SIGINT
	waitGroup.Wait()
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	stdoutWriter.Wait()
	stderrWriter.Wait()
	require.Equal(t, "\nExecution cancelled...\n", stdoutWriter.String())
	require.Equal(t, "Failed to cancel: 2 errors occurred:\n\t* cancelling a run that has not yet started\n\t* test error\n\n\n", stderrWriter.String())
}

func TestExecutionContext_CollectErrors(t *testing.T) {
	pipedWriteCloser := customio.NewPipedWriteCloser()
	executionContext := NewExecutionContext()
	run := executionContext.FullRun(
		WithLogWriter(pipedWriteCloser),
		WithTearDownFunc(func(run *pipeline.Run) {
			run.Log.Error(fmt.Errorf("test error"))
		}),
	)
	run.Wait()
	pipedWriteCloser.Wait()
	require.Equal(t, 1, executionContext.errors.Len())
	require.Equal(t, "anonymous:\ntest error", executionContext.errors.Errors[0].Error())
}

func TestExecutionContext_WaitForRun(t *testing.T) {
	executionContext := NewExecutionContext()
	runIdentifier := "test"
	var run *pipeline.Run
	go func() {
		time.Sleep(500)
		executionContext.FullRun(
			WithIdentifier(&runIdentifier),
			WithSetupFunc(func(calledRun *pipeline.Run) {
				run = calledRun
			}),
		)
	}()
	waitedRun := executionContext.WaitForRun("test")
	require.Equal(t, run, waitedRun)
}

func TestExecutionContext_UserRuns(t *testing.T) {
	executionContext := NewExecutionContext()
	test0 := "test0"
	test1 := "test1"
	test2 := "test2"
	executionContext.runs = []*pipeline.Run{
		{
			Identifier: &test0,
			Definition: &pipeline.Definition{
				BuiltIn: true,
			},
		},
		{
			Identifier: nil,
			Definition: &pipeline.Definition{
				BuiltIn: true,
			},
		},
		{
			Identifier: &test1,
			Definition: nil,
		},
		{
			Identifier: &test2,
			Definition: &pipeline.Definition{
				BuiltIn: false,
			},
		},
		{
			Identifier: nil,
			Definition: &pipeline.Definition{
				BuiltIn: false,
			},
		},
	}
	require.Equal(t, []*pipeline.Run{
		{
			Identifier: &test1,
			Definition: nil,
		},
		{
			Identifier: &test2,
			Definition: &pipeline.Definition{
				BuiltIn: false,
			},
		},
		{
			Identifier: nil,
			Definition: &pipeline.Definition{
				BuiltIn: false,
			},
		},
	}, executionContext.UserRuns())
}

type FakeMiddleware struct {
	CallCount int
}

func NewFakeMiddleware() *FakeMiddleware {
	return &FakeMiddleware{
		CallCount: 0,
	}
}

func (fakeMiddleware *FakeMiddleware) String() string {
	return "fake"
}

func (fakeMiddleware *FakeMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *ExecutionContext,
) {
	fakeMiddleware.CallCount++
	next(run)
}
