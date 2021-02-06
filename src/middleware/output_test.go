package middleware

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/datastream"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRun_Output_startProgress(t *testing.T) {
	buffer := new(bytes.Buffer)
	executionContext := NewExecutionContext(
		WithActivityIndicator(
			logging.NewNestedActivityIndicator(),
		),
	)
	startProgress(executionContext, buffer)
	require.Contains(t, buffer.String(), fmt.Sprintln("==== PROGRESS ===="))
}

func TestRun_Output_stopProgress(t *testing.T) {
	executionContext := NewExecutionContext(
		WithActivityIndicator(
			logging.NewNestedActivityIndicator(),
		),
	)
	require.True(t, executionContext.ActivityIndicator.Visible())
	stopProgress(executionContext)
	require.False(t, executionContext.ActivityIndicator.Visible())
}

func TestRun_Output_outputResult(t *testing.T) {
	buffer := new(bytes.Buffer)
	outputResult(&pipeline.Run{
		Stdout: datastream.NewClosedComposableDataStreamFromBuffer(bytes.NewBuffer([]byte("test"))),
	}, buffer)
	require.Equal(t, fmt.Sprintln("===== RESULT =====")+fmt.Sprintln("test"), buffer.String())
}

func TestRun_Output_outputResult_nil(t *testing.T) {
	buffer := new(bytes.Buffer)
	outputResult(nil, buffer)
	require.Equal(t, fmt.Sprintln("===== RESULT =====")+fmt.Sprintln(aurora.Gray(12, "no result")), buffer.String())
}

func TestRun_outputLogs(t *testing.T) {
	buffer := new(bytes.Buffer)
	outputLogs(&pipeline.Run{
		Log: pipeline.NewClosedPipelineRunLoggerWithResult(bytes.NewBuffer([]byte("test"))),
	}, buffer)
	require.Equal(t, fmt.Sprintln("====== LOGS ======")+fmt.Sprintln("test"), buffer.String())
}

func TestRun_Output_outputErrors(t *testing.T) {
	buffer := new(bytes.Buffer)
	outputErrors(&multierror.Error{
		Errors: []error{
			fmt.Errorf("test error 1"),
			fmt.Errorf("test error 2"),
		},
	}, buffer)
	require.Equal(t, fmt.Sprintln("===== ERRORS =====")+fmt.Sprintln("test error 1")+fmt.Sprintln("test error 2"), buffer.String())
}
