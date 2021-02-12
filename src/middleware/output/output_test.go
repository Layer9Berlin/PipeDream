package _output

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "↗️️ output | overwritten output")
	require.Equal(t, "overwritten output", run.Stdout.String())
}
