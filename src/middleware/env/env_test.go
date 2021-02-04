package env

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"pipedream/src/models"
	"strings"
	"testing"
)

func TestEnv_InterpolateShallow(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	env := map[string]string{
		"TEST": "new",
	}
	setenv := func(key string, value string) error {
		env[key] = value
		return nil
	}
	expandEnv := func(value string) string {
		result := value
		for key, newValue := range env {
			result = strings.Replace(value, "$"+key, newValue, -1)
		}
		return result
	}
	var calledRun *models.PipelineRun = nil
	NewEnvMiddlewareWithProvider(setenv, expandEnv).Apply(
		run,
		func(pipelineRun *models.PipelineRun) {
			calledRun = pipelineRun
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.NotNil(t, calledRun)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow new",
	}, calledRun.ArgumentsCopy())
	require.Contains(t, run.Log.String(), "env")
}

func TestEnv_InterpolateDeep(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "deep",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	env := map[string]string{
		"TEST": "new",
	}
	setenv := func(key string, value string) error {
		env[key] = value
		return nil
	}
	expandEnv := func(value string) string {
		result := value
		for key, newValue := range env {
			result = strings.Replace(value, "$"+key, newValue, -1)
		}
		return result
	}
	var calledRun *models.PipelineRun = nil
	NewEnvMiddlewareWithProvider(setenv, expandEnv).Apply(
		run,
		func(pipelineRun *models.PipelineRun) {
			calledRun = pipelineRun
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.NotNil(t, calledRun)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "deep",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "new",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another new",
				}},
			}},
			"yet another new",
		},
		"test2": "shallow new",
	}, calledRun.ArgumentsCopy())
	require.Contains(t, run.Log.String(), "env")
}

func TestEnv_NoSubstitutions(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "deep",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	env := map[string]string{
		"NOT_PRESENT": "new",
	}
	setenv := func(key string, value string) error {
		env[key] = value
		return nil
	}
	expandEnv := func(value string) string {
		result := value
		for key, newValue := range env {
			result = strings.Replace(value, "$"+key, newValue, -1)
		}
		return result
	}
	var calledRun *models.PipelineRun = nil
	NewEnvMiddlewareWithProvider(setenv, expandEnv).Apply(
		run,
		func(pipelineRun *models.PipelineRun) {
			calledRun = pipelineRun
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.NotNil(t, calledRun)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "deep",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, calledRun.ArgumentsCopy())
	require.NotContains(t, run.Log.String(), "env")
}

func TestEnv_InterpolateNone(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "none",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	env := map[string]string{
		"TEST": "new",
	}
	setenv := func(key string, value string) error {
		env[key] = value
		return nil
	}
	expandEnv := func(value string) string {
		result := value
		for key, newValue := range env {
			result = strings.Replace(value, "$"+key, newValue, -1)
		}
		return result
	}
	var calledRun *models.PipelineRun = nil
	NewEnvMiddlewareWithProvider(setenv, expandEnv).Apply(
		run,
		func(pipelineRun *models.PipelineRun) {
			calledRun = pipelineRun
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.NotNil(t, calledRun)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"env": map[string]interface{}{
			"interpolate": "none",
		},
		"test": []interface{}{
			map[string]interface{}{"test1": map[string]interface{}{
				"value": "$TEST",
			}},
			map[string]interface{}{"test2": map[string]interface{}{
				"test": map[string]interface{}{"test2": map[string]interface{}{
					"value": "another $TEST",
				}},
			}},
			"yet another $TEST",
		},
		"test2": "shallow $TEST",
	}, calledRun.ArgumentsCopy())
	require.NotContains(t, run.Log.String(), "env")
}

func TestEnv_Save(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"env": map[string]interface{}{
			"save": "KEY",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	env := map[string]string{
		"TEST": "new",
	}
	setenv := func(key string, value string) error {
		env[key] = value
		return nil
	}
	expandEnv := func(value string) string {
		result := value
		for key, newValue := range env {
			result = strings.Replace(value, "$"+key, newValue, -1)
		}
		return result
	}
	var calledRun *models.PipelineRun = nil
	NewEnvMiddlewareWithProvider(setenv, expandEnv).Apply(
		run,
		func(pipelineRun *models.PipelineRun) {
			calledRun = pipelineRun
			pipelineRun.Stdout.Replace(strings.NewReader("test"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.NotNil(t, calledRun)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test", run.Stdout.String())
	require.Equal(t, "test", env["KEY"])
	require.Contains(t, run.Log.String(), "env")
}

func TestEnv_NewEnvMiddleware(t *testing.T) {
	envMiddleware := NewEnvMiddleware()
	key := "PIPEDREAM_TEST_TEMP"
	err := envMiddleware.Setenv(key, "value")
	require.Nil(t, err)
	require.Equal(t, "value", os.Getenv(key))
	require.Equal(t, "test value", envMiddleware.ExpandEnv("test $"+key))
	_ = os.Unsetenv(key)
	require.Equal(t, "", os.Getenv(key))
	require.Equal(t, "test ", envMiddleware.ExpandEnv("test $"+key))
}
