package each

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestEach_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"each": []interface{}{
			map[string]interface{}{"pipe1": map[string]interface{}{
				"arg": "value",
			}},
			map[string]interface{}{"pipe2": map[string]interface{}{
				"arg": "value",
			}},
			map[*string]interface{}{nil: map[string]interface{}{
				"arg": "value",
			}},
		},
	}, nil, nil)

	waitGroup := &sync.WaitGroup{}
	allInputs := make([]string, 0, 3)
	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdin.Replace(strings.NewReader("bla\nbla\nend\n"))
	run.Stdout.Replace(strings.NewReader("output of parent pipe\n"))
	NewMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(
				func(childRun *pipeline.Run) {
					stdinCopy := childRun.Stdin.Copy()
					waitGroup.Add(1)
					go func() {
						defer waitGroup.Done()
						completeInput, err := ioutil.ReadAll(stdinCopy)
						require.Nil(t, err)
						allInputs = append(allInputs, string(completeInput))
					}()
					identifier := "anonymous"
					if childRun.Identifier != nil {
						identifier = *childRun.Identifier
					}
					childRun.Stdout.Replace(strings.NewReader(fmt.Sprintf("output of pipeline `%v`\n", identifier)))
				}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, []string{"bla\nbla\nend\n", "bla\nbla\nend\n", "bla\nbla\nend\n"}, allInputs)
	require.Equal(t, "output of parent pipe\noutput of pipeline `pipe1`\noutput of pipeline `pipe2`\noutput of pipeline `anonymous`\n", run.Stdout.String())
	logString := run.Log.String()
	require.Contains(t, logString, "each")
	require.Contains(t, logString, "pipe1, pipe2, ~")
}

func TestEach_ApplyWithInvalidArguments(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"each": []interface{}{
			map[string]interface{}{"pipe1": interface{}(
				"invalid",
			)},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\nbla\nend\n"))
			run.Stdout.Replace(strings.NewReader("output\n"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments for \"each\"")
	require.Equal(t, "bla\nbla\nend\n", run.Stdin.String())
	require.Equal(t, "output\n", run.Stdout.String())
}

func TestEach_Inactive(t *testing.T) {
	run, _ := pipeline.NewRun(nil, nil, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\nbla\nend\n"))
			run.Stdout.Replace(strings.NewReader("output of parent pipe\n"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "bla\nbla\nend\n", run.Stdin.String())
	require.Equal(t, "output of parent pipe\n", run.Stdout.String())
	require.Contains(t, run.Log.String(), "↘️12B")
}
