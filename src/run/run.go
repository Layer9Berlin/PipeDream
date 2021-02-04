package run

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/middleware/middleware_stack"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"path/filepath"
)

var Log = logrus.New()
var Verbosity string
var PipelineFlag string
var FileFlag string

var executionContextFactory = middleware.NewExecutionContext
var osStdin io.ReadCloser = os.Stdin
var osStdout io.WriteCloser = os.Stdout

func Cmd(_ *cobra.Command, args []string) {
	executableLocation, _ := os.Executable()
	executableDir := path.Dir(executableLocation)
	projectPath, _ := filepath.EvalSymlinks(executableDir)
	executionContext := executionContextFactory(
		//middleware.WithActivityIndicator(logging.NewBarActivityIndicator(osStdout)),
		middleware.WithMiddlewareStack(middleware_stack.SetUpMiddleware()),
		middleware.WithProjectPath(projectPath),
		middleware.WithLogger(Log),
	)
	err := executionContext.SetUpPipelines(args)
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

	executionContext.Execute(pipelineIdentifier, osStdout)
}
