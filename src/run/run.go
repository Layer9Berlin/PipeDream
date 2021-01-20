package run

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"pipedream/src/logging"
	"pipedream/src/middleware"
	"pipedream/src/middleware/middleware_stack"
)

var Log = logrus.New()
var Verbosity string

var executionContextFactory = middleware.NewExecutionContext
var osStdin io.ReadCloser = os.Stdin
var osStdout io.WriteCloser = os.Stdout

func Cmd(_ *cobra.Command, args []string) {
	projectPath, _ := os.Getwd()

	executionContext := executionContextFactory(
		middleware.WithActivityIndicator(logging.NewNestedActivityIndicator()),
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
