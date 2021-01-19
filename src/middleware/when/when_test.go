package when

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"pipedream/src/models"
	"testing"
)

func TestWhen_TrueCondition(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"when": "8 in (7,8,9)",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewWhenMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "when | satisfied | \"8 in (7,8,9)\"")
}

func TestWhen_FalseCondition(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"when": "1 == 2",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewWhenMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "when | not satisfied | \"1 == 2\"")
}

func TestWhen_UnparseableCondition(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"when": "1 == '",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewWhenMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "error parsing condition")
	require.Contains(t, run.Log.String(), "error parsing condition \"1 == '\"")
}

func TestWhen_WithoutCondition(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewWhenMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.NotContains(t, run.Log.String(), "when")
}
