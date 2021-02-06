package with

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestWith_Pattern(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"with": map[string]interface{}{
			"pattern": "(?m)^test.*",
		},
	}, nil, nil)

	waitGroup := &sync.WaitGroup{}
	allInputs := make([]string, 0, 3)
	NewWithMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\ntest1\nbla\ntest2\ntest3\nend\n"))
		},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				stdinCopier := childRun.Stdin.Copy()
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					completeInput, err := ioutil.ReadAll(stdinCopier)
					require.Nil(t, err)
					allInputs = append(allInputs, string(completeInput))
				}()
				childRun.Stdout.Replace(strings.NewReader("child output"))
				childRun.Stderr.Replace(strings.NewReader("child stderr\n"))
			}),
		))
	waitGroup.Wait()
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "bla\nchild output\nbla\nchild output\nchild output\nend\n", run.Stdout.String())
	require.Equal(t, "child stderr\nchild stderr\nchild stderr\n", run.Stderr.String())
}

func TestWith_NoPattern(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, nil, nil, nil)

	NewWithMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\ntest1\nbla\ntest2\ntest3\nend\n"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "bla\ntest1\nbla\ntest2\ntest3\nend\n", run.Stdin.String())
	require.Equal(t, "", run.Stdout.String())
	require.Equal(t, "", run.Stderr.String())
}

func TestWith_NoMatch(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"with": map[string]interface{}{
			"pattern": "(?m)^test.*",
		},
	}, nil, nil)

	NewWithMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\nbla\nend\n"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "bla\nbla\nend\n", run.Stdout.String())
	require.Equal(t, "", run.Stderr.String())
}

func TestWith_PatternDoesNotCompile(t *testing.T) {
	run, _ := pipeline.NewPipelineRun(nil, map[string]interface{}{
		"with": map[string]interface{}{
			"pattern": "(?m^test.*",
		},
	}, nil, nil)

	NewWithMiddleware().Apply(
		run,
		func(pipelineRun *pipeline.Run) {
			run.Stdin.Replace(strings.NewReader("bla\nbla\nend\n"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "error parsing regexp")
	require.Equal(t, "bla\nbla\nend\n", run.Stdin.String())
	require.Equal(t, "", run.Stdout.String())
	require.Equal(t, "", run.Stderr.String())
}
