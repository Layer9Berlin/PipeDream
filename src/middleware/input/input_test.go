package _input

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInput_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"input": map[string]interface{}{
			"text": "overwritten input",
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
	require.Contains(t, run.Log.String(), "↘️ input | overwritten input")
	require.Equal(t, "overwritten input", run.Stdin.String())
}
