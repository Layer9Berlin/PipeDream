package pipe

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestPipe_Apply(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"pipe": []interface{}{
			map[string]interface{}{
				"test1": map[string]interface{}{
					"arg": "value",
				},
			},
			"test2",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.InfoLevel)
	waitGroup := &sync.WaitGroup{}
	run.Stdin.Replace(strings.NewReader("test input"))
	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				switch *childRun.Identifier {
				case "test1":
					stdinCopy := childRun.Stdin.Copy()
					waitGroup.Add(1)
					go func() {
						defer waitGroup.Done()
						completeInput, err := ioutil.ReadAll(stdinCopy)
						require.Nil(t, err)
						require.Equal(t, "test input", string(completeInput))
					}()
					childRun.Stdout.Replace(strings.NewReader("test1 output"))
					childRun.Log.Info(fields.Message("test1 log entry"))
				case "test2":
					stdinCopy := childRun.Stdin.Copy()
					waitGroup.Add(1)
					go func() {
						completeInput, err := ioutil.ReadAll(stdinCopy)
						require.Nil(t, err)
						require.Equal(t, "test1 output", string(completeInput))
						waitGroup.Done()
					}()
					childRun.Stdout.Replace(strings.NewReader("test2 output"))
					childRun.Log.Info(fields.Message("test2 log entry"))
				}
			}),
		))
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test2 output", run.Stdout.String())
	logString := run.Log.String()
	require.Contains(t, logString, "test1 log entry")
	require.Contains(t, logString, "test2 log entry")
}

func TestPipe_InvalidArguments(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"pipe": "invalid",
	}, nil, nil)

	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {
			run.Stdout.Replace(strings.NewReader("test output"))
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
	require.Equal(t, "test output", run.Stdout.String())
}

func TestPipe_NotInvoked(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, nil, nil, nil)

	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {
			run.Stdout.Replace(strings.NewReader("test output"))
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test output", run.Stdout.String())
}

func TestPipe_InvalidReference(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"pipe": []interface{}{
			map[string]interface{}{
				"test1": map[string]interface{}{
					"arg": "value",
				},
				"this key": map[string]interface{}{
					"makes": "the reference invalid",
				},
			},
		},
	}, nil, nil)

	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {
			run.Stdout.Replace(strings.NewReader("test output"))
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "invalid pipeline reference")
	require.Equal(t, "test output", run.Stdout.String())
}

func TestPipe_AnonymousReference(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"pipe": []interface{}{
			map[*string]interface{}{
				nil: map[string]interface{}{
					"arg": "value",
				},
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	var fullRunCalled = false
	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {
			run.Stdout.Replace(strings.NewReader("test output"))
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				fullRunCalled = true
			}),
		))
	run.Start()
	run.Wait()

	require.True(t, fullRunCalled)
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test output", run.Stdout.String())
	require.Contains(t, run.Log.String(), "anonymous")
}
