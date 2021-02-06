package catcheach

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCatchEach_Error(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"catchEach": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	handlerCalled := false
	NewCatchEachMiddleware().Apply(run, func(run *pipeline.Run) {
		run.Stdout.Replace(strings.NewReader("output\n"))
		run.Stderr.Replace(strings.NewReader("test error"))
	}, middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(errorRun *pipeline.Run) {
			require.Equal(t, "test-handler", *errorRun.Identifier)
			handlerCalled = true
			errorRun.Stdout.Replace(strings.NewReader("handled"))
		})))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.True(t, handlerCalled)
	require.Equal(t, "output\nhandled", run.Stdout.String())
	require.Contains(t, run.Log.String(), "catchEach")
}

func TestCatchEach_MultipleLinesOfErrorOutput(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"catchEach": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	handlerInvocations := 0
	NewCatchEachMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = io.WriteString(stdoutIntercept, "output\n")
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
		stderrIntercept := run.Stderr.Intercept()
		go func() {
			_, err := io.Copy(stderrIntercept, stderrIntercept)
			require.Nil(t, err)
			_, err = io.WriteString(stderrIntercept, "test error\nanother error\nyet another error")
			require.Nil(t, err)
			require.Nil(t, stderrIntercept.Close())
		}()
	}, middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(errorRun *pipeline.Run) {
			handlerInvocations += 1
			stdoutWriter := errorRun.Stdout.WriteCloser()
			go func() {
				_, err := io.WriteString(stdoutWriter, "handled\n")
				require.Nil(t, err)
				require.Nil(t, stdoutWriter.Close())
			}()
		})))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 3, handlerInvocations)
	require.Equal(t, "output\nhandled\nhandled\nhandled\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "catchEach")
}

func TestCatchEach_HandlerNotInvoked(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"catchEach": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewCatchEachMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = stdoutIntercept.Write([]byte("output\n"))
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
	}, nil)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "output\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "catchEach")
}

func TestCatchEach_NoCatchHandler(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewCatchEachMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = stdoutIntercept.Write([]byte("output\n"))
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
	}, nil)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "output\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "↗️7B")
}

func TestCatchEach_HandlerThrowingError(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"catchEach": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	handlerCalled := false
	NewCatchEachMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = stdoutIntercept.Write([]byte("output\n"))
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
		stderrIntercept := run.Stderr.Intercept()
		go func() {
			_, err := io.Copy(stderrIntercept, stderrIntercept)
			require.Nil(t, err)
			_, err = stderrIntercept.Write([]byte("test error"))
			require.Nil(t, err)
			require.Nil(t, stderrIntercept.Close())
		}()
	}, middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(errorRun *pipeline.Run) {
			require.Equal(t, "test-handler", *errorRun.Identifier)
			handlerCalled = true
			errorRun.Stdout.Replace(strings.NewReader("not properly handled"))
			errorRun.Stderr.Replace(strings.NewReader("handler error"))
		})))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.True(t, handlerCalled)
	require.Equal(t, "output\nnot properly handled", run.Stdout.String())
	require.Equal(t, "handler error", run.Stderr.String())
	require.Contains(t, run.Log.String(), "catchEach")
}
