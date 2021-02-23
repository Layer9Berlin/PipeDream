package _switch

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestSwitch_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"switch": []interface{}{
			map[string]string{
				"pattern": "test1",
				"text":    "match1",
			},
			map[string]string{
				"pattern": "test2",
				"text":    "match2",
			},
			map[string]string{
				"text": "default",
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdin.Replace(strings.NewReader("this string\nwill match test1\n"))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "🔢 switch | match")
	require.NotContains(t, logString, "🔢 switch | mismatch")
	require.Equal(t, "match1", run.Stdout.String())
}

func TestSwitch_Apply_matchSecondPattern(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"switch": []interface{}{
			map[string]string{
				"pattern": "test1",
				"text":    "match1",
			},
			map[string]string{
				"pattern": "test2",
				"text":    "match2",
			},
			map[string]string{
				"text": "default",
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdin.Replace(strings.NewReader("this string\nwill match test2\n"))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "🔢 switch | mismatch")
	require.Contains(t, logString, "🔢 switch | match")
	require.Equal(t, "match2", run.Stdout.String())
}

func TestSwitch_Apply_matchDefault(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"switch": []interface{}{
			map[string]string{
				"pattern": "test1",
				"text":    "match1",
			},
			map[string]string{
				"pattern": "test2",
				"text":    "match2",
			},
			map[string]string{
				"text": "default",
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdin.Replace(strings.NewReader("this string\nwill not match any pattern\n"))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.Contains(t, logString, "🔢 switch | mismatch")
	require.NotContains(t, logString, "🔢 switch | match | test")
	require.Contains(t, logString, "🔢 switch | match | default")
	require.Equal(t, "default", run.Stdout.String())
}

func TestSwitch_Apply_missingText(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"switch": []interface{}{
			map[string]string{
				"pattern": "test1",
			},
			map[string]string{
				"text": "default",
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	run.Stdin.Replace(strings.NewReader("this string\nwill match test1\n"))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	logString := run.Log.String()
	require.NotContains(t, logString, "🔢 switch | match | test")
	require.Contains(t, logString, "🔢 switch | match | default")
	require.Equal(t, "default", run.Stdout.String())
}
