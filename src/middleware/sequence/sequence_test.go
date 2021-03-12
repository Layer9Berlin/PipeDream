package sequence

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestSequence_Apply(t *testing.T) {
	executionSequence := make([]string, 0, 16)
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"sequence": []string{
			"test1",
			"test2",
			"test3",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	executionContext := middleware.NewExecutionContext(
		middleware.WithExecutionFunction(
			func(childRun *pipeline.Run) {
				childRun.Stdout.Replace(strings.NewReader("ok\n"))
				executionSequence = append(executionSequence, "start "+childRun.Name())
				time.Sleep(100 * time.Millisecond)
				executionSequence = append(executionSequence, "stop "+childRun.Name())
			},
		))
	NewMiddleware().Apply(
		run,
		func(childRun *pipeline.Run) {},
		executionContext,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "wait for completion | test1")
	require.Contains(t, logString, "wait for completion | test2")
	require.Equal(t, "ok\nok\nok\n", run.Stdout.String())
	require.Equal(t, []string{"start test1", "stop test1", "start test2", "stop test2", "start test3", "stop test3"}, executionSequence)

}
