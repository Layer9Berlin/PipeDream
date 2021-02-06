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
	dirMiddleware := DirMiddleware{
		DirChanger: func(newDir string) error {
			currentDir = newDir
			return nil
		},
	}
	currentDir = "pre-test"
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"dir": "changed",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	dirMiddleware.Apply(
		run,
		func(invocation *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, "changed", currentDir)
	require.Contains(t, run.Log.String(), "dir")
}

func TestDir_DontChangeDir(t *testing.T) {
	dirMiddleware := DirMiddleware{
		DirChanger: func(newDir string) error {
			currentDir = newDir
			return nil
		},
	}
	currentDir = "pre-test"
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	dirMiddleware.Apply(
		run,
		func(invocation *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, "pre-test", currentDir)
	require.NotContains(t, run.Log.String(), "dir")
}

func TestDir_ErrorChangingDir(t *testing.T) {
	dirMiddleware := DirMiddleware{
		DirChanger: func(newDir string) error {
			return fmt.Errorf("error changing directory")
		},
	}

	currentDir = "pre-test"
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"dir": "changed",
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	dirMiddleware.Apply(
		run,
		func(invocation *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, "pre-test", currentDir)
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Equal(t, "error changing directory", run.Log.LastError().Error())
	require.Contains(t, run.Log.String(), "dir")
}

func TestDir_CreateNewDirMiddleware(t *testing.T) {
	dirMiddleware := NewDirMiddleware()
	require.NotNil(t, dirMiddleware)
	require.NotNil(t, dirMiddleware.DirChanger)
}
