package pipeline

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/logrusorgru/aurora/v3"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPipelineRun_AppendToStdout(t *testing.T) {
	run, setupErr := NewRun(nil, nil, nil, nil)
	require.Nil(t, setupErr)

	run.Stdout.MergeWith(strings.NewReader("this is a test"))
	run.Start()
	run.Wait()

	require.Equal(t, "this is a test", run.Stdout.String())
}

func TestPipelineRun_ArgumentAtPath(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key": "value",
		},
	}, nil, nil)
	argument, err := run.ArgumentAtPath("test", "key")
	require.Nil(t, err)
	argumentAsString, argumentIsString := argument.(string)
	require.True(t, argumentIsString)
	require.Equal(t, "value", argumentAsString)
}

func TestPipelineRun_ArgumentAtPathIncludingParents(t *testing.T) {
	parentRun, _ := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key": "value",
		},
	}, nil, nil)
	run, _ := NewRun(nil, nil, nil, parentRun)
	argument, err := run.ArgumentAtPathIncludingParents("test", "key")
	require.Nil(t, err)
	argumentAsString, argumentIsString := argument.(string)
	require.True(t, argumentIsString)
	require.Equal(t, "value", argumentAsString)
}

func TestPipelineRun_ArgumentAtPathIncludingParents_NotFound(t *testing.T) {
	parentRun, _ := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key": "value",
		},
	}, nil, nil)
	run, _ := NewRun(nil, nil, nil, parentRun)
	argument, err := run.ArgumentAtPathIncludingParents("test", "key", "not_present")
	require.NotNil(t, err)
	require.Equal(t, "value does not exist at path", err.Error())
	require.Nil(t, argument)
}

func TestPipelineRun_SetArgumentAtPath(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key": "value",
		},
	}, nil, nil)
	err := run.SetArgumentAtPath("new_value", "test", "new_key")
	require.Nil(t, err)
	require.Equal(t, map[string]interface{}{
		"test": map[string]interface{}{
			"key":     "value",
			"new_key": "new_value",
		},
	}, run.ArgumentsCopy())
}

func TestPipelineRun_HaveArgumentAtPath(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key": "value",
		},
		"test2": []string{
			"test3",
			"test4",
		},
	}, nil, nil)
	require.False(t, run.HaveArgumentAtPath("missing"))
	require.False(t, run.HaveArgumentAtPath("missing", "as", "well"))
	require.False(t, run.HaveArgumentAtPath("test", "missing"))
	require.False(t, run.HaveArgumentAtPath("test2", "test3"))
	require.True(t, run.HaveArgumentAtPath("test2"))
	require.True(t, run.HaveArgumentAtPath("test", "key"))
}

func TestPipelineRun_Lengths(t *testing.T) {
	run, _ := NewRun(nil, nil, nil, nil)
	run.Stdin.MergeWith(strings.NewReader("test"))
	run.Stdout.MergeWith(strings.NewReader("another test"))
	run.Start()
	run.Wait()
	require.Equal(t, 4, run.Stdin.Len())
	require.Equal(t, 12, run.Stdout.Len())
	require.Equal(t, 0, run.Stderr.Len())
}

func TestPipelineRun_NewPipelineRun(t *testing.T) {
	definition := NewDefinition(map[string]interface{}{
		"test": "value",
	}, "test", true, false)
	run, _ := NewRun(nil, nil, definition, nil)
	require.Equal(t, map[string]interface{}{
		"test": "value",
	}, run.ArgumentsCopy())
}

func TestPipelineRun_WaitForCompletion(t *testing.T) {
	run, _ := NewRun(nil, nil, nil, nil)
	appender := run.Stderr.WriteCloser()
	// Wait should be callable an arbitrary number of times simultaneously
	go func() {
		run.Wait()
	}()
	go func() {
		run.Wait()
	}()
	go func() {
		run.Wait()
	}()
	go func() {
		_, _ = appender.Write([]byte("test"))
		_ = appender.Close()
	}()
	run.Start()
	run.Wait()
	require.Equal(t, "test", run.Stderr.String())
}

func TestPipelineRun_String(t *testing.T) {
	identifier := "test"
	run, _ := NewRun(&identifier, nil, nil, nil)
	run.Stdin.MergeWith(strings.NewReader("test"))
	run.Stdout.MergeWith(strings.NewReader("test output"))
	run.Stderr.MergeWith(strings.NewReader("test err"))
	run.Log.Error(fmt.Errorf("test error"))
	run.Log.Warn(fields.Message("test warning"))
	run.Start()
	run.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Bold("Test"), "  ",
		aurora.Gray(12, "↘️4B"), "  ",
		aurora.Gray(12, "↗️11B"), "  ",
		aurora.Red("⛔️8B"), "  ",
		aurora.Yellow("⚠️1"), "  ",
		aurora.Red("🛑1"),
	), run.String())
}

func TestPipelineRun_String_WithDescription(t *testing.T) {
	identifier := "test"
	run, _ := NewRun(&identifier, map[string]interface{}{
		"description": "Test description",
	}, nil, nil)
	run.Stdin.MergeWith(strings.NewReader("test"))
	run.Stdout.MergeWith(strings.NewReader("test output"))
	run.Stderr.MergeWith(strings.NewReader("test err"))
	run.Log.Error(fmt.Errorf("test error"))
	run.Log.Warn(fields.Message("test warning"))
	run.Start()
	run.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Bold("Test description"), "  ",
		aurora.Gray(12, "↘️4B"), "  ",
		aurora.Gray(12, "↗️11B"), "  ",
		aurora.Red("⛔️8B"), "  ",
		aurora.Yellow("⚠️1"), "  ",
		aurora.Red("🛑1"),
	), run.String())
}

func TestPipelineRun_UnmergeableDefinition(t *testing.T) {
	definition := NewDefinition(map[string]interface{}{
		"test": "value",
	}, "test", true, false)
	run, err := NewRun(nil, map[string]interface{}{
		"test": map[string]interface{}{
			"key1": "value1",
		},
	}, definition, nil)
	require.Nil(t, run)
	require.NotNil(t, err)
}

func TestPipelineRun_SetArguments(t *testing.T) {
	run, err := NewRun(nil, map[string]interface{}{}, nil, nil)
	require.NotNil(t, run)
	require.Nil(t, err)
	run.SetArguments(map[string]interface{}{
		"test": map[string]interface{}{
			"test1": "test2",
		},
	})
	value, err := run.ArgumentAtPath("test", "test1")
	require.Nil(t, err)
	require.Equal(t, "test2", value)
}

func TestPipelineRun_Close(t *testing.T) {
	run, _ := NewRun(nil, nil, nil, nil)

	require.False(t, run.Completed())

	// closing several times is allowed
	run.Start()
	run.Start()
	run.Start()

	run.Wait()

	require.True(t, run.Completed())
}

func TestPipelineRun_RemoveArgumentAtPath(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test1": map[string]interface{}{
			"test2": map[string]interface{}{
				"test3": "test4",
			},
			"test5": "test6",
		},
	}, nil, nil)

	err := run.RemoveArgumentAtPath("test1", "test2")

	require.Nil(t, err)
	require.Equal(t, map[string]interface{}{
		"test1": map[string]interface{}{
			"test5": "test6",
		},
	}, run.arguments)
}

func TestPipelineRun_Cancel(t *testing.T) {
	run, _ := NewRun(nil, nil, nil, nil)
	run.AddCancelHook(func() error {
		return fmt.Errorf("test error")
	})

	require.False(t, run.Cancelled())

	err := run.Cancel()

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")

	require.True(t, run.Cancelled())
}

func TestPipelineRun_GraphLabel(t *testing.T) {
	runIdentifier := "test"
	run, _ := NewRun(&runIdentifier, nil, nil, nil)
	run.executionWaitGroup.Add(1)
	require.Equal(t, "🔜 Test", run.GraphLabel())
	run.Start()
	require.Equal(t, "↺ Test", run.GraphLabel())
	run.executionWaitGroup.Done()
	run.Wait()
	require.Equal(t, "✔ Test", run.GraphLabel())
	_ = run.Cancel()
	require.Equal(t, "⎋ Test", run.GraphLabel())
	run.Log.Error(fmt.Errorf("test error"))
	require.Equal(t, "✘ Test", run.GraphLabel())
}

func TestPipelineRun_GraphGroup(t *testing.T) {
	runIdentifier := "test"
	run, _ := NewRun(&runIdentifier, nil, nil, nil)
	run.executionWaitGroup.Add(1)
	require.Equal(t, "waiting", run.GraphGroup())
	run.Start()
	require.Equal(t, "active", run.GraphGroup())
	run.executionWaitGroup.Done()
	run.Wait()
	require.Equal(t, "success", run.GraphGroup())
	_ = run.Cancel()
	require.Equal(t, "cancelled", run.GraphGroup())
	run.Log.Error(fmt.Errorf("test error"))
	require.Equal(t, "error", run.GraphGroup())
}

func TestPipelineRun_Name(t *testing.T) {
	anonymousRun, _ := NewRun(nil, nil, nil, nil)
	require.Equal(t, "anonymous", anonymousRun.Name())
	runIdentifier := "test"
	namedRun, _ := NewRun(&runIdentifier, nil, nil, nil)
	require.Equal(t, "test", namedRun.Name())
}
