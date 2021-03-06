package catch

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

func TestCatch_Error(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"catch": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	handlerCalled := false
	NewMiddleware().Apply(run, func(run *pipeline.Run) {
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
			handlerCalled = true
			require.Equal(t, "test-handler", *errorRun.Identifier)
			errorRun.Stdout.Replace(strings.NewReader("handled"))
		})))
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.True(t, handlerCalled)
	require.Equal(t, "output\nhandled", run.Stdout.String())
}

func TestCatch_MultipleLinesOfErrorOutput(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"catch": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	handlerInvocations := 0
	NewMiddleware().Apply(run, func(run *pipeline.Run) {
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
			_, err = stderrIntercept.Write([]byte("test error\nanother error\nyet another error"))
			require.Nil(t, err)
			require.Nil(t, stderrIntercept.Close())
		}()
	}, middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(errorRun *pipeline.Run) {
			handlerInvocations++
			require.Equal(t, "test-handler", *errorRun.Identifier)
			errorRun.Stdout.Replace(strings.NewReader("handled"))
		})))
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, handlerInvocations)
	require.Equal(t, "output\nhandled", run.Stdout.String())
	require.Contains(t, run.Log.String(), "catch")
}

func TestCatch_HandlerNotInvoked(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"catch": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = stdoutIntercept.Write([]byte("output\n"))
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
	}, nil)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "output\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "catch")
}

func TestCatch_WithoutHandler(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(run, func(run *pipeline.Run) {
		stdoutIntercept := run.Stdout.Intercept()
		go func() {
			_, err := ioutil.ReadAll(stdoutIntercept)
			require.Nil(t, err)
			_, err = stdoutIntercept.Write([]byte("output\n"))
			require.Nil(t, err)
			require.Nil(t, stdoutIntercept.Close())
		}()
	}, nil)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "output\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "↗️7B")
}

func TestCatch_HandlerThrowingError(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"catch": "test-handler",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	handlerCalled := false
	NewMiddleware().Apply(run, func(run *pipeline.Run) {
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
		}()
		go func() {
			_, err := stderrIntercept.Write([]byte("test error"))
			require.Nil(t, err)
			require.Nil(t, stderrIntercept.Close())
		}()
	}, middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(errorRun *pipeline.Run) {
			handlerCalled = true
			require.Equal(t, "test-handler", *errorRun.Identifier)
			errorRun.Stdout.Replace(strings.NewReader("not properly handled"))
			errorRun.Stderr.Replace(strings.NewReader("handler error"))
		})))
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.True(t, handlerCalled)
	require.Equal(t, "output\nnot properly handled", run.Stdout.String())
	require.Equal(t, "handler error", run.Stderr.String())
	require.Contains(t, run.Log.String(), "catch")
}
