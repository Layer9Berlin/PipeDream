package _select

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestSelect_Apply(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"select": map[string]interface{}{
			"prompt":  "test prompt",
			"initial": 2,
			"options": []interface{}{
				"test1",
				interface{}(map[string]interface{}{
					"test2": map[string]interface{}{},
				}),
				interface{}(map[string]interface{}{
					"test3": map[string]interface{}{
						"description": "description3",
					},
				}),
				interface{}(map[*string]interface{}{
					nil: map[string]interface{}{},
				}),
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)

	executionContext := middleware.NewExecutionContext(middleware.WithUserPromptImplementation(
		func(
			label string,
			items []string,
			initialSelection int,
			size int,
			input io.ReadCloser,
			output io.WriteCloser,
		) (int, string, error) {
			require.Equal(t, "test prompt", label)
			require.Equal(t, []string{"test1", "test2", "description3", "-"}, items)
			return 1, "", nil
		}),
		middleware.WithExecutionFunction(
			func(childRun *pipeline.Run) {
				if *childRun.Identifier == "test2" {
					childRun.Stdout.MergeWith(childRun.Stdin.Copy())
				}
			},
		))
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		executionContext,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "👈 select | user selected pipeline | test2")
}

func TestSelect_StdinAndStdout(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"select": map[string]interface{}{
			"prompt":  "test prompt",
			"initial": 2,
			"options": []interface{}{
				"test1",
				interface{}(map[string]interface{}{
					"test2": map[string]interface{}{},
				}),
				interface{}(map[string]interface{}{
					"test3": map[string]interface{}{
						"description": "description3",
					},
				}),
				interface{}(map[*string]interface{}{
					nil: map[string]interface{}{},
				}),
			},
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)

	runStdIn := ioutil.NopCloser(new(bytes.Buffer))
	runStdoutReader, runStdOut := io.Pipe()
	executionContext := middleware.NewExecutionContext(middleware.WithUserPromptImplementation(
		func(
			label string,
			items []string,
			initialSelection int,
			size int,
			input io.ReadCloser,
			output io.WriteCloser,
		) (int, string, error) {
			require.Equal(t, "test prompt", label)
			require.Equal(t, []string{"test1", "test2", "description3", "-"}, items)
			return 1, "", nil
		}),
		middleware.WithExecutionFunction(
			func(childRun *pipeline.Run) {
				if *childRun.Identifier == "test2" {
					childRun.Stdout.MergeWith(childRun.Stdin.Copy())
					go func() {
						childRun.Wait()
						runStdOut.Close()
					}()
				}
			},
		))
	run.Stdin.Replace(strings.NewReader("test stdin"))
	NewMiddlewareWithStdinAndStdout(runStdIn, runStdOut).Apply(
		run,
		func(run *pipeline.Run) {},
		executionContext,
	)
	output := []byte{}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		output, _ = ioutil.ReadAll(runStdoutReader)
		waitGroup.Done()
	}()
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "👈 select | user selected pipeline | test2")
	require.Contains(t, run.Stdout.String(), "test stdin")
	require.Equal(t, "test stdin", string(output))
}
