// Package collect provides a middleware for gathering values into a map and saving them to disk
package collect

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Middleware is a directory navigator
type Middleware struct {
	fileWriter func(fileName string, data string) error
	WorkingDir string
}

// String is a human-readable description
func (Middleware) String() string {
	return "collect"
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware() Middleware {
	workingDir, _ := os.Getwd()
	return Middleware{
		fileWriter: defaultWriteToFile,
		WorkingDir: workingDir,
	}
}

type middlewareArguments struct {
	File   *string
	Nested *bool
	Values []pipeline.Reference
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (collectMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	collectArguments := middlewareArguments{}
	pipeline.ParseArguments(&collectArguments, "collect", run)

	if len(collectArguments.Values) > 0 {
		childIdentifiers, childArguments, info := pipeline.CollectReferences(collectArguments.Values)
		fileName := "-"
		if collectArguments.File != nil {
			fileName = filepath.Base(*collectArguments.File)
		}
		run.Log.Debug(
			fields.Symbol("🧺"),
			fields.Message(fileName),
			fields.Info(strings.Join(info, ", ")),
			fields.Middleware(collectMiddleware),
		)

		waitGroup := sync.WaitGroup{}
		results := make(map[string]interface{}, len(collectArguments.Values))
		resultsMutex := sync.Mutex{}
		for index, childIdentifier := range childIdentifiers {
			arguments := childArguments[index]
			identifier := childIdentifier
			if identifier != nil {
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(identifier),
					middleware.WithArguments(arguments),
					middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
						run.Log.Trace(
							fields.DataStream(collectMiddleware, "copy child stdout into parent stdout")...,
						)
						childStdout := childRun.Stdout.Copy()
						waitGroup.Add(1)
						childRun.DontCompleteBefore(func() {
							defer waitGroup.Done()
							completeChildStdout, err := ioutil.ReadAll(childStdout)
							childRun.Log.PossibleError(err)
							resultsMutex.Lock()
							defer resultsMutex.Unlock()
							if collectArguments.Nested != nil && *collectArguments.Nested {
								var childResult interface{}
								err = yaml.Unmarshal(completeChildStdout, &childResult)
								childRun.Log.PossibleError(err)
								results[*identifier] = childResult
							} else {
								results[*identifier] = string(completeChildStdout)
							}
						})
					}))
			}
		}
		stdoutWriteCloser := run.Stdout.WriteCloser()
		run.DontCompleteBefore(func() {
			waitGroup.Wait()
			yamlResults, err := yaml.Marshal(results)
			run.Log.PossibleError(err)
			if collectArguments.File == nil {
				_, err = stdoutWriteCloser.Write(yamlResults)
				run.Log.PossibleError(err)
			} else {
				err = collectMiddleware.fileWriter(
					filepath.Join(collectMiddleware.WorkingDir, *collectArguments.File),
					string(yamlResults),
				)
				run.Log.PossibleError(err)
			}
			_ = stdoutWriteCloser.Close()
		})
	}

	next(run)
}

func defaultWriteToFile(fileName string, data string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}
