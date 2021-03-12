package dir

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

var currentDir = "/test"

func TestDir_ChangeDir(t *testing.T) {
	dirMiddleware := Middleware{
		DirChanger: func(newDir string) error {
			currentDir = newDir
			return nil
		},
		WorkingDir: "working-dir",
	}
	currentDir = "pre-test"
	runDir := ""
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"dir": "changed",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	dirMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			runDir = currentDir
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, "working-dir", currentDir)
	require.Equal(t, "changed", runDir)
	require.Contains(t, run.Log.String(), "dir")
}

func TestDir_DontChangeDir(t *testing.T) {
	dirMiddleware := Middleware{
		DirChanger: func(newDir string) error {
			currentDir = newDir
			return nil
		},
		WorkingDir: "working-dir",
	}
	currentDir = "pre-test"
	runDir := ""
	run, _ := pipeline.NewRun(nil, map[string]interface{}{}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	dirMiddleware.Apply(
		run,
		func(run *pipeline.Run) {
			runDir = currentDir
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, "working-dir", currentDir)
	require.Equal(t, "pre-test", runDir)
	require.NotContains(t, run.Log.String(), "dir")
}

func TestDir_ErrorChangingDir(t *testing.T) {
	dirMiddleware := Middleware{
		DirChanger: func(newDir string) error {
			return fmt.Errorf("error changing directory")
		},
		WorkingDir: "working-dir",
	}

	currentDir = "pre-test"
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"dir": "changed",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	runDir := ""
	dirMiddleware.Apply(
		run,
		func(invocation *pipeline.Run) {
			runDir = currentDir
		},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, "pre-test", currentDir)
	require.Equal(t, "pre-test", runDir)
	require.Equal(t, 2, run.Log.ErrorCount())
	require.Equal(t, "error changing directory", run.Log.LastError().Error())
	require.Contains(t, run.Log.String(), "dir")
}

func TestDir_CreateNewDirMiddleware(t *testing.T) {
	dirMiddleware := NewMiddleware()
	require.NotNil(t, dirMiddleware)
	require.NotNil(t, dirMiddleware.DirChanger)
}
