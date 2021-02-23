package run

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/graph"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/parsing"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

func TestRun_Cmd_noBuiltInPipelineFiles(t *testing.T) {
	oldExecutionContextFactory := executionContextFactory
	defer func() {
		executionContextFactory = oldExecutionContextFactory
	}()
	buffer := new(bytes.Buffer)
	executionContextFactory = func(options ...middleware.ExecutionContextOption) *middleware.ExecutionContext {
		options = append(options, middleware.WithParser(parsing.NewParser(parsing.WithFindByGlobImplementation(func(_ string) ([]string, error) {
			return []string{}, nil
		}))))
		executionContext := middleware.NewExecutionContext(options...)
		executionContext.Log.SetOutput(buffer)
		return executionContext
	}

	Cmd(nil, []string{})

	require.Contains(t, buffer.String(), "please double-check your installation")
}

func TestRun_Cmd_userPromptError(t *testing.T) {
	oldExecutionContextFactory := executionContextFactory
	defer func() {
		executionContextFactory = oldExecutionContextFactory
	}()
	buffer := new(bytes.Buffer)
	executionContextFactory = func(options ...middleware.ExecutionContextOption) *middleware.ExecutionContext {
		options = append(options, middleware.WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
				parsing.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte{}, nil
				}),
				parsing.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return paths, nil
				}),
			)),
			middleware.WithUserPromptImplementation(func(
				label string,
				items []string,
				initialSelection int,
				size int,
				input io.ReadCloser,
				output io.WriteCloser,
			) (int, string, error) {
				return 0, "", fmt.Errorf("test error")
			}),
		)
		executionContext := middleware.NewExecutionContext(options...)
		executionContext.Log.SetOutput(buffer)
		return executionContext
	}

	Cmd(nil, []string{})

	require.Contains(t, buffer.String(), "test error")
}

func TestRun_Cmd(t *testing.T) {
	oldStdout := osStdout
	reader, writer := io.Pipe()
	defer func() {
		osStdout = oldStdout
	}()
	osStdout = writer
	oldExecutionContextFactory := executionContextFactory
	defer func() {
		executionContextFactory = oldExecutionContextFactory
	}()
	buffer := new(bytes.Buffer)
	result := make([]byte, 0, 1024)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		result, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	executionContextFactory = func(options ...middleware.ExecutionContextOption) *middleware.ExecutionContext {
		options = append(options, middleware.WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(_ string) ([]string, error) {
					return []string{"test1.pipe"}, nil
				}),
				parsing.WithReadFileImplementation(func(_ string) ([]byte, error) {
					return []byte(`
public:
  test:
    arg: value
`), nil
				}),
				parsing.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return []string{"test1.pipe"}, nil
				}),
			)))
		executionContext := middleware.NewExecutionContext(options...)
		executionContext.Log.SetOutput(buffer)
		return executionContext
	}

	Cmd(nil, []string{"test1.pipe"})

	_ = writer.Close()
	waitGroup.Wait()
	require.Contains(t, string(result), "===== RESULT =====")
	require.NotContains(t, string(result), "====== LOGS ======")
	require.Equal(t, "", buffer.String())
}

func TestRun_Cmd_withGraphFlag(t *testing.T) {
	ShowGraphFlag = true
	previousGraphWriter := graphWriter
	defer func() {
		ShowGraphFlag = false
		graphWriter = previousGraphWriter
		if recover() == nil {
			t.Errorf("failed to encounter expected panic")
		}
	}()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	graphWriter = graph.NewWriter()
	graphWriter.OpenInBrowser = func(file string) error {
		waitGroup.Done()
		return fmt.Errorf("test error")
	}

	oldExecutionContextFactory := executionContextFactory
	defer func() {
		executionContextFactory = oldExecutionContextFactory
	}()
	executionContextFactory = func(options ...middleware.ExecutionContextOption) *middleware.ExecutionContext {
		options = append(options, middleware.WithParser(
			parsing.NewParser(
				parsing.WithFindByGlobImplementation(func(_ string) ([]string, error) {
					return []string{"test1.pipe"}, nil
				}),
				parsing.WithReadFileImplementation(func(_ string) ([]byte, error) {
					return []byte(`
public:
  test:
    arg: value
`), nil
				}),
				parsing.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return []string{"test1.pipe"}, nil
				}),
			)))
		executionContext := middleware.NewExecutionContext(options...)
		return executionContext
	}

	Cmd(nil, []string{"test1.pipe"})
}
