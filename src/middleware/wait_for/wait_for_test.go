package waitformiddleware

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWaitFor_Pipes(t *testing.T) {
	runIdentifier1 := "test1"
	runIdentifier2 := "test2"
	run1, _ := pipeline.NewRun(&runIdentifier1, map[string]interface{}{}, nil, nil)
	run2, _ := pipeline.NewRun(&runIdentifier2, map[string]interface{}{
		"waitFor": map[string]interface{}{
			"pipes": []string{"test1"},
		},
	}, nil, nil)

	run1.Log.SetLevel(logrus.TraceLevel)
	run2.Log.SetLevel(logrus.TraceLevel)

	executionContext := middleware.NewExecutionContext()
	run1WriteCloser := run1.Stdout.WriteCloser()
	NewMiddleware().Apply(
		run1,
		func(run *pipeline.Run) {},
		executionContext,
	)
	run1.Close()
	executionContext.Runs = []*pipeline.Run{run1}
	NewMiddleware().Apply(
		run2,
		func(run *pipeline.Run) {},
		executionContext,
	)
	run2.Close()
	require.False(t, run2.Completed())

	time.Sleep(100)

	run1WriteCloser.Close()

	run1.Wait()
	run2.Wait()

	require.True(t, run2.Completed())

	require.Equal(t, 0, run1.Log.ErrorCount())
	require.Equal(t, 0, run2.Log.ErrorCount())
	require.Contains(t, run2.Log.String(), "🕙 waitFor | waiting for run \"test1\"")
}

func TestWaitFor_EnvVars(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"waitFor": map[string]interface{}{
			"envVars": []string{"test"},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)

	envVars := make(map[string]interface{}, 1)
	NewMiddlewareWithLookupImplementation(func(key string) (string, bool) {
		result, ok := envVars[key].(string)
		return result, ok
	}).Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()

	time.Sleep(100)

	require.False(t, run.Completed())

	envVars["test"] = "value"

	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "🕙 waitFor | waiting for env var \"test\" to be set")
}
