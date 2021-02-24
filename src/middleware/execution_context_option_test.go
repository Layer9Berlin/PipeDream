package middleware

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"
)

func TestExecutionContext_WithProjectPath(t *testing.T) {
	executionContext := NewExecutionContext(
		WithProjectPath("test/path"),
	)
	require.Equal(t, "test/path", executionContext.ProjectPath)
}

func TestExecutionContext_WithMiddlewareStack(t *testing.T) {
	middleware1 := NewFakeMiddleware()
	middleware2 := NewFakeMiddleware()
	stack := []Middleware{
		middleware1,
		middleware2,
	}
	executionContext := NewExecutionContext(
		WithMiddlewareStack(stack),
	)
	require.Equal(t, stack, executionContext.MiddlewareStack)
}

func TestExecutionContext_WithLogger(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := logrus.New()
	logger.SetOutput(buffer)
	executionContext := NewExecutionContext(
		WithLogger(logger),
	)
	executionContext.Log.Error("test")
	require.Contains(t, buffer.String(), "test")
}

func TestExecutionContext_WithUserPromptImplementation(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := logrus.New()
	logger.SetOutput(buffer)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	executionContext := NewExecutionContext(
		WithUserPromptImplementation(func(
			label string,
			items []string,
			initialSelection int,
			size int,
			input io.ReadCloser,
			output io.WriteCloser,
		) (int, string, error) {
			waitGroup.Done()
			return 0, "", nil
		}),
	)
	_, _, _ = executionContext.UserPromptImplementation("test", nil, 0, 0, nil, nil)
	waitGroup.Wait()
}
