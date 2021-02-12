// Package run contains the implementation of the default `pipedream` shell command, selecting, parsing and executing pipelines
package run

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/middleware/stack"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"path/filepath"
)

// Log is the main logger that all other loggers are based on
var Log = logrus.New()

// Verbosity determines the global log levels
//
// It corresponds to the log level of user-defined pipelines.
// Note that built-in pipelines may have a different log level.
var Verbosity string

// PipelineFlag sets the pipeline to be executed, skipping the user selection prompt
var PipelineFlag string

// FileFlag sets the file to be executed, skipping the user selection prompt
var FileFlag string

var executionContextFactory = middleware.NewExecutionContext
var osStdin io.ReadCloser = os.Stdin
var osStdout io.WriteCloser = os.Stdout
var osStderr io.WriteCloser = os.Stderr

// Cmd executes the main command, selecting and running a pipeline within an execution context
func Cmd(_ *cobra.Command, args []string) {
	executableLocation, _ := os.Executable()
	executableDir := path.Dir(executableLocation)
	projectPath, _ := filepath.EvalSymlinks(executableDir)
	executionContext := executionContextFactory(
		middleware.WithMiddlewareStack(stack.SetUpMiddleware()),
		middleware.WithProjectPath(projectPath),
		middleware.WithLogger(Log),
	)
	err := executionContext.SetUpPipelines(FileFlag, args)
	if err != nil {
		executionContext.Log.Error(err)
		return
	}

	pipelineIdentifier, fileName, err := letUserSelectPipelineFileAndPipeline(executionContext, 10, osStdin, osStdout)
	if err != nil {
		executionContext.Log.Error(err)
		return
	}
	executionContext.RootFileName = fileName

	executionContext.Execute(pipelineIdentifier, osStdout, osStderr)
}
