package inherit

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInherit_ArgumentsFromParent(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"arg":       "value",
		"arg2":      "another-value",
		"arg-array": []interface{}{"line0", "line1", "line2"},
		"arg-num":   13,
		"arg3": map[string]interface{}{
			"test1": map[string]interface{}{
				"test": []interface{}{
					"one",
					"two",
					"three",
				},
			},
			"test2": "test3",
		},
		"arg4": map[string]interface{}{
			"test": "test",
		},
	}, nil, nil)

	childRun, _ := pipeline.NewRun(nil, map[string]interface{}{
		"inherit": []interface{}{
			"arg",
			"arg-array",
			"arg-num",
			"arg3",
			"arg4",
		},
		"arg4": map[string]interface{}{
			"existing": "value",
		},
	}, nil, run)

	childRun.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		childRun,
		func(invocation *pipeline.Run) {},
		nil,
	)
	childRun.Start()
	childRun.Wait()

	require.Equal(t, map[string]interface{}{
		"arg":       "value",
		"arg-array": []interface{}{"line0", "line1", "line2"},
		"arg-num":   13,
		"arg3": map[string]interface{}{
			"test1": map[string]interface{}{
				"test": []interface{}{
					"one",
					"two",
					"three",
				},
			},
			"test2": "test3",
		},
		"arg4": map[string]interface{}{
			"existing": "value",
		},
		"inherit": []interface{}{
			"arg",
			"arg-array",
			"arg-num",
			"arg3",
			"arg4",
		},
	}, childRun.ArgumentsCopy())

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, childRun.Log.String(), "inherit")
}

func TestInherit_WithoutParent(t *testing.T) {
	identifier := "orphan"
	orphanRun, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"inherit": []interface{}{
			"arg",
			"arg-array",
			"arg-num",
			"arg3",
			"arg4",
		},
		"arg4": map[string]interface{}{
			"existing": "value",
		},
	}, nil, nil)

	orphanRun.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		orphanRun,
		func(invocation *pipeline.Run) {},
		nil,
	)
	orphanRun.Start()
	orphanRun.Wait()

	require.Equal(t, map[string]interface{}{
		"inherit": []interface{}{
			"arg",
			"arg-array",
			"arg-num",
			"arg3",
			"arg4",
		},
		"arg4": map[string]interface{}{
			"existing": "value",
		},
	}, orphanRun.ArgumentsCopy())
	require.Equal(t, 0, orphanRun.Log.ErrorCount())
	require.Equal(t, fmt.Sprint(
		aurora.Green(fmt.Sprint("▶️ starting | ", aurora.Bold("Orphan"))),
		"\n",
		aurora.Green(fmt.Sprint("✔ completed | ", aurora.Bold("Orphan"))),
		"\n",
	), orphanRun.Log.String())
}

func TestInherit_NoValue(t *testing.T) {
	parentRun, _ := pipeline.NewRun(nil, map[string]interface{}{}, nil, nil)

	childIdentifier := "child"
	childRun, _ := pipeline.NewRun(&childIdentifier, map[string]interface{}{
		"inherit": []interface{}{
			"arg",
		},
	}, nil, parentRun)

	childRun.Log.SetLevel(logrus.TraceLevel)
	NewMiddleware().Apply(
		childRun,
		func(invocation *pipeline.Run) {},
		nil,
	)
	childRun.Start()
	childRun.Wait()

	require.Equal(t, 0, childRun.Log.ErrorCount())
	require.Equal(t, fmt.Sprint(
		aurora.Green(fmt.Sprint("  ▶️ starting | ", aurora.Bold("Child"))),
		"\n",
		aurora.Green(fmt.Sprint("  ✔ completed | ", aurora.Bold("Child"))),
		"\n",
	), childRun.Log.String())
}
