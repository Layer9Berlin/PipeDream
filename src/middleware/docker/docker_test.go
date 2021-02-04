package docker

import (
	"github.com/Layer9Berlin/pipedream/src/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDocker_RunInDockerContainer(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"docker": map[string]interface{}{
			"service": "test-service",
		},
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewDockerMiddleware().Apply(
		run,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "docker-compose exec -T test-service test", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "docker")
}

func TestDocker_MalformedArgument(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"docker": map[string]interface{}{
			"test": "test-service",
		},
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	NewDockerMiddleware().Apply(
		run,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments for \"docker\"")
	require.Equal(t, "test", run.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, run.Log.String(), "malformed arguments for \"docker\"")
}

func TestDocker_InheritArgument(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"docker": map[string]interface{}{
			"service": "test-service",
		},
	}, nil, nil)

	childRun, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "test",
		},
	}, nil, run)

	childRun.Log.SetLevel(logrus.TraceLevel)
	NewDockerMiddleware().Apply(
		childRun,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	childRun.Close()
	childRun.Wait()

	require.Equal(t, 0, childRun.Log.ErrorCount())
	require.Equal(t, "docker-compose exec -T test-service test", childRun.ArgumentsCopy()["shell"].(map[string]interface{})["run"].(string))
	require.Contains(t, childRun.Log.String(), "docker")
}

func TestDocker_NonRunnable(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"docker": map[string]interface{}{
			"service": "test-service",
		},
	}, nil, nil)

	childRun, _ := models.NewPipelineRun(nil, map[string]interface{}{}, nil, run)

	childRun.Log.SetLevel(logrus.TraceLevel)
	NewDockerMiddleware().Apply(
		childRun,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	childRun.Close()
	childRun.Wait()

	require.Equal(t, 0, childRun.Log.ErrorCount())
	require.Contains(t, childRun.Log.String(), "docker")
}
