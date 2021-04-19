// Package extract provides a middleware merging values from nested maps into pipe arguments
package extract

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Middleware is a directory navigator
type Middleware struct {
	fileReader func(fileName string) ([]byte, error)
	WorkingDir string
}

// String is a human-readable description
func (Middleware) String() string {
	return "extract"
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware() Middleware {
	workingDir, _ := os.Getwd()
	return Middleware{
		fileReader: ioutil.ReadFile,
		WorkingDir: workingDir,
	}
}

type middlewareArguments struct {
	File   string
	Values map[string][]string
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (extractMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	collectArguments := middlewareArguments{}
	pipeline.ParseArguments(&collectArguments, "extract", run)

	if len(collectArguments.Values) > 0 {
		fileName := filepath.Base(collectArguments.File)
		run.Log.Debug(
			fields.Symbol("🔍"),
			fields.Message(fileName),
			fields.Info(fmt.Sprintf("%v values", len(collectArguments.Values))),
			fields.Middleware(extractMiddleware),
		)

		fileData, err := extractMiddleware.fileReader(collectArguments.File)
		run.Log.PossibleError(err)
		fileDataAsYaml := make(map[string]interface{}, 10)
		err = yaml.Unmarshal(fileData, &fileDataAsYaml)
		run.Log.PossibleError(err)
		for key, valuePath := range collectArguments.Values {
			value, err := stringmap.GetValueInMap(fileDataAsYaml, valuePath...)
			if err != nil {
				run.Log.PossibleError(fmt.Errorf("failed to find value at path `%v`: %w", strings.Join(valuePath, "."), err))
			}
			if valueAsString, valueIsString := value.(string); valueIsString {
				run.Log.Trace(
					fields.Symbol("🔍"),
					fields.Message(fileName),
					fields.Info(fmt.Sprintf("%q: %q", key, valueAsString)),
					fields.Middleware(extractMiddleware),
				)
				err = run.SetArgumentAtPath(valueAsString, key)
				run.Log.PossibleError(err)
			} else {
				run.Log.Error(fmt.Errorf("extracted value not a string, but of unexpected type: `%T`", value))
			}
		}
	}

	next(run)
}
