package sync_middleware

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"testing"
	"time"
)

func TestSyncMiddleware_Apply(t *testing.T) {
	identifier := "test"
	run, _ := models.NewPipelineRun(&identifier, map[string]interface{}{
		"sync": []interface{}{
			map[string]interface{}{
				"test1": map[string]interface{}{
					"arg": "value",
				},
			},
			"test2",
		},
	}, nil, nil)

	executionOrder := make([]string, 0, 4)
	run.Log.SetLevel(logrus.DebugLevel)
	NewSyncMiddleware().Apply(run,
		func(run *models.PipelineRun) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *models.PipelineRun) {
				switch *childRun.Identifier {
				case "test1":
					executionOrder = append(executionOrder, "test1 start")
					stdoutWriter := childRun.Stdout.WriteCloser()
					go func() {
						time.Sleep(300)
						_, _ = io.WriteString(stdoutWriter, "test1 output")
						executionOrder = append(executionOrder, "test1 end")
						_ = stdoutWriter.Close()
					}()
				case "test2":
					executionOrder = append(executionOrder, "test2 start")
					stdoutWriter := childRun.Stdout.WriteCloser()
					go func() {
						_, _ = io.WriteString(stdoutWriter, "test2 output")
						executionOrder = append(executionOrder, "test2 end")
						_ = stdoutWriter.Close()
					}()
				}
			}),
		))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test2 output", run.Stdout.String())
	require.Contains(t, run.Log.String(), "sync")
}

func TestSyncMiddleware_SingleAnonymousPipe(t *testing.T) {
	identifier := "test"
	run, _ := models.NewPipelineRun(&identifier, map[string]interface{}{
		"sync": []interface{}{
			map[*string]interface{}{
				nil: map[string]interface{}{
					"arg": "value",
				},
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewSyncMiddleware().Apply(run,
		func(run *models.PipelineRun) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *models.PipelineRun) {
				stdoutWriter := childRun.Stdout.WriteCloser()
				go func() {
					_, _ = io.WriteString(stdoutWriter, "test output")
					_ = stdoutWriter.Close()
				}()
			}),
		))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test output", run.Stdout.String())
	require.Contains(t, run.Log.String(), "sync")
}

func TestSyncMiddleware_NoInvocation(t *testing.T) {
	identifier := "test"
	run, _ := models.NewPipelineRun(&identifier, map[string]interface{}{
		"sync": []interface{}{
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewSyncMiddleware().Apply(run,
		func(run *models.PipelineRun) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *models.PipelineRun) {
			}),
		))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "", run.Stdout.String())
	require.NotContains(t, run.Log.String(), "sync")
}
