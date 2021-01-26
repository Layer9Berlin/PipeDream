package middleware

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"io"
	"pipedream/src/models"
	"strings"
)

func startProgress(executionContext *ExecutionContext, writer io.Writer) {
	if executionContext.ActivityIndicator != nil {
		_, _ = fmt.Fprintln(writer, "==== PROGRESS ====")
	}
}

func stopProgress(executionContext *ExecutionContext) {
	if executionContext.ActivityIndicator != nil {
		executionContext.ActivityIndicator.Wait()
	}
}

func outputResult(result *models.PipelineRun, writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "===== RESULT =====")
	if result == nil || result.Stdout.Len() == 0 {
		_, _ = fmt.Fprintln(writer, aurora.Gray(12, "no result"))
		return
	} else {
		_, _ = fmt.Fprintln(writer, result.Stdout.String())
	}
}

func outputLogs(run *models.PipelineRun, writer io.Writer) {
	if run != nil && run.Log != nil {
		logOutput := run.Log.String()
		if len(logOutput) > 0 {
			_, _ = fmt.Fprintln(writer, "====== LOGS ======")
			_, _ = fmt.Fprintln(writer, logOutput)
		}
	}
}

func outputErrors(run *models.PipelineRun, writer io.Writer) {
	if run.Log.ErrorCount() > 0 {
		_, _ = fmt.Fprintln(writer, "===== ERRORS =====")
		_, _ = fmt.Fprintln(writer, strings.Join(run.Log.AllErrorMessages(), "\n"))
	}
}
