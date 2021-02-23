package when

import (
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWhen_TrueCondition(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"when": "8 in (7,8,9)",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "when | satisfied | \"8 in (7,8,9)\"")
}

func TestWhen_FalseCondition(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"when": "1 == 2",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "when | not satisfied | \"1 == 2\"")
}

func TestWhen_UnparseableCondition(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"when": "1 == '",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "error parsing condition")
	require.Contains(t, run.Log.String(), "error parsing condition \"1 == '\"")
}

func TestWhen_WithoutCondition(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.NotContains(t, run.Log.String(), "when")
}

func TestWhen_else_withStringReference(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"when": "false",
		"else": "test",
	}, nil, nil)

	childIdentifier := "test"
	runArguments := make(stringmap.StringMap, 1)
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				childRun.Log.Debug(fields.Message("child run log entry"))
				childIdentifier = *childRun.Identifier
				runArguments = childRun.ArgumentsCopy()
			}),
		),
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "✖️ when | else")
	require.Equal(t, "test", childIdentifier)
	require.Equal(t, map[string]interface{}{}, runArguments)
}

func TestWhen_else_withMapReference(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"when": "false",
		"else": map[string]interface{}{
			"test": map[string]interface{}{
				"arg": "value",
			},
		},
	}, nil, nil)

	childIdentifier := "test"
	runArguments := make(stringmap.StringMap, 1)
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				childRun.Log.Debug(fields.Message("child run log entry"))
				childIdentifier = *childRun.Identifier
				runArguments = childRun.ArgumentsCopy()
			}),
		),
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "✖️ when | else")
	require.Equal(t, "test", childIdentifier)
	require.Equal(t, map[string]interface{}{
		"arg": "value",
	}, runArguments)
}
