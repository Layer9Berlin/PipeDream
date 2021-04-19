package collect

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestCollect_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"collect": map[string]interface{}{
			"values": []interface{}{
				map[string]interface{}{"pipe1": map[string]interface{}{
					"arg": "value",
				}},
				map[string]interface{}{"pipe2": map[string]interface{}{
					"arg": "value",
				}},
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(
				func(childRun *pipeline.Run) {
					childRun.Stdout.Replace(strings.NewReader(fmt.Sprintf("result of pipeline `%v`\n", *childRun.Identifier)))
				}),
		))
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "pipe1: |\n  result of pipeline `pipe1`\npipe2: |\n  result of pipeline `pipe2`\n", run.Stdout.String())
	logString := run.Log.String()
	require.Contains(t, logString, "collect")
	require.Contains(t, logString, "pipe1, pipe2")
}

func TestCollect_ApplyWithFile(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"collect": map[string]interface{}{
			"file": "test.yaml",
			"values": []interface{}{
				map[string]interface{}{"pipe1": map[string]interface{}{
					"arg": "value",
				}},
				map[string]interface{}{"pipe2": map[string]interface{}{
					"arg": "value",
				}},
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	filePath := ""
	result := ""
	Middleware{fileWriter: func(file string, data string) error {
		filePath = file
		result = data
		return nil
	},
		WorkingDir: "test-dir",
	}.Apply(
		run,
		func(pipelineRun *pipeline.Run) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(
				func(childRun *pipeline.Run) {
					childRun.Stdout.Replace(strings.NewReader(fmt.Sprintf("result of pipeline `%v`\n", *childRun.Identifier)))
				}),
		))
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "pipe1: |\n  result of pipeline `pipe1`\npipe2: |\n  result of pipeline `pipe2`\n", result)
	require.Equal(t, "", run.Stdout.String())
	require.Equal(t, "test-dir/test.yaml", filePath)

	logString := run.Log.String()
	require.Contains(t, logString, "collect")
	require.Contains(t, logString, "pipe1, pipe2")
}
