package _output

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"testing"
)

func TestOutput_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"output": map[string]interface{}{
			"text": "overwritten output",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "↗️️ output | overwritten output")
	require.Equal(t, "overwritten output", run.Stdout.String())
}

func TestOutput_Process_stringReference(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"output": map[string]interface{}{
			"process": "output-processor",
		},
	}, nil, nil)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdout.Replace(strings.NewReader("test output"))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(run *pipeline.Run) {
				go func() {
					run.Wait()
					require.Equal(t, run.Stdin.String(), "test output")
					waitGroup.Done()
				}()
				require.Equal(t, "output-processor", *run.Identifier)
				run.Stdout.Replace(strings.NewReader("processor output"))
			})),
	)
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "⍈ output | process")
	require.Equal(t, "processor output", run.Stdout.String())
}
