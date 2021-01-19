package ssh

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"pipedream/src/models"
	"testing"
)

func TestRunningViaSsh(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"ssh": map[string]interface{}{
			"host": "test-host",
		},
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewSshMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "ssh test-host \"bash -l -c \\\"test\\\"\"", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "ssh")
}

func TestNestedPipelines(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"ssh": map[string]interface{}{
			"host": "test-host",
		},
	}, nil, nil)

	childRun, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, run)

	childRun.Log.SetLevel(logrus.TraceLevel)
	NewSshMiddleware().Apply(
		childRun,
		func(run *models.PipelineRun) {},
		nil,
	)
	childRun.Close()
	childRun.Wait()

	require.Equal(t, 0, childRun.Log.ErrorCount())
	require.Equal(t, "ssh test-host \"bash -l -c \\\"test\\\"\"", childRun.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, childRun.Log.String(), "ssh")
}

func TestMissingHostArgument(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"ssh": map[string]interface{}{
			"test": "not-a-host-arg",
		},
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewSshMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments for \"ssh\"")
	require.Equal(t, "test", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "malformed arguments")
}

func TestNilArgument(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewSshMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "anonymous")
}

func TestMalformedArgument(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"ssh": map[string]interface{}{
			"test": []interface{}{
				"invalid",
			},
		},
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewSshMiddleware().Apply(
		run,
		func(run *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments for \"ssh\"")
	require.Equal(t, "test", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "malformed arguments")
}
