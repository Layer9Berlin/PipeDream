package inherit

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"pipedream/src/models"
	"testing"
)

func TestInherit_ArgumentsFromParent(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
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

	childRun, _ := models.NewPipelineRun(nil, map[string]interface{}{
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
	NewInheritMiddleware().Apply(
		childRun,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	childRun.Close()
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
	orphanRun, _ := models.NewPipelineRun(&identifier, map[string]interface{}{
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
	NewInheritMiddleware().Apply(
		orphanRun,
		func(invocation *models.PipelineRun) {},
		nil,
	)
	orphanRun.Close()
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
		aurora.Gray(18, fmt.Sprint("⏏️ closing | ", aurora.Bold("Orphan"))),
		"\n",
		aurora.Green(fmt.Sprint("✔ completed | ", aurora.Bold("Orphan"))),
		"\n",
	), orphanRun.Log.String())
}
