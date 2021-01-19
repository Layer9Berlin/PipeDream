package models

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"pipedream/src/logging/log_fields"
	"testing"
)

func TestPipelineRunLogger_PossibleError(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)
	logger.PossibleError(nil)
	logger.PossibleError(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
}

func TestPipelineRunLogger_Error(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)
	logger.Error(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
	require.Equal(t, "test error", logger.LastError().Error())
	logger.Error(fmt.Errorf("another error"), log_fields.Middleware("test 2"), )
	require.Equal(t, 2, logger.ErrorCount())
	require.Equal(t, "another error", logger.LastError().Error())
}

func TestPipelineRunLogger_Counts(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	require.Equal(t, 0, logger.TraceCount())
	logger.TraceWithFields()
	require.Equal(t, 1, logger.TraceCount())

	require.Equal(t, 0, logger.DebugCount())
	logger.DebugWithFields()
	require.Equal(t, 1, logger.DebugCount())

	require.Equal(t, 0, logger.InfoCount())
	logger.InfoWithFields()
	require.Equal(t, 1, logger.InfoCount())

	require.Equal(t, 0, logger.WarnCount())
	logger.WarnWithFields()
	require.Equal(t, 1, logger.WarnCount())

	require.Equal(t, 0, logger.ErrorCount())
	logger.Error(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
}

func TestPipelineRunLogger_Close(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	require.False(t, logger.Closed())
	require.False(t, logger.Completed())

	logger.Close()

	require.True(t, logger.Closed())

	logger.Wait()

	require.True(t, logger.Completed())
}

func TestPipelineRunLogger_PossibleErrorWithExplanation(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.PossibleErrorWithExplanation(nil, "just testing: ")
	require.Equal(t, 0, logger.ErrorCount())

	logger.PossibleErrorWithExplanation(fmt.Errorf("test error"), "just testing:")
	require.Equal(t, 1, logger.ErrorCount())
	require.Equal(t, "just testing: test error", logger.LastError().Error())
}

func TestPipelineRunLogger_LastError(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	require.Nil(t, logger.LastError())
	err1 := fmt.Errorf("test error 1")
	logger.Error(err1)
	require.Equal(t, err1, logger.LastError())
	err2 := fmt.Errorf("test error 2")
	logger.Error(err2)
	require.Equal(t, err2, logger.LastError())
}

func TestPipelineRunLogger_AllErrorMessages(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	err1 := fmt.Errorf("test error 1")
	logger.Error(err1)
	err2 := fmt.Errorf("test error 2")
	logger.Error(err2)
	require.Equal(t, []string{"test error 1", "test error 2"}, logger.AllErrorMessages())
}

func TestPipelineRunLogger_CloseAlreadyClosed(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	// closing multiple times is supported
	logger.Close()
	logger.Close()
	logger.Close()

	require.True(t, logger.Closed())
}

func TestPipelineRunLogger_Output(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.SetLevel(logrus.TraceLevel)
	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	), logger.String())
	require.Equal(t, []byte(fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	)), logger.Bytes())
}

func TestPipelineRunLogger_WriteToReader_TraceLevel(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.SetLevel(logrus.TraceLevel)
	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_DebugLevel(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.SetLevel(logrus.DebugLevel)
	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_InfoLevel(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.SetLevel(logrus.InfoLevel)
	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_WarningLevel(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.SetLevel(logrus.WarnLevel)
	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_ErrorLevel(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(logrus.WithField("message", "test warning"))
	logger.Info(logrus.WithField("message", "test info"))
	logger.Debug(logrus.WithField("message", "test debug"))
	logger.Trace(logrus.WithField("message", "test trace"))

	logger.SetLevel(logrus.ErrorLevel)
	logger.Close()
	logger.Wait()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_NestedReaders(t *testing.T) {
	logger := NewPipelineRunLogger(nil, 0)
	logger2 := NewPipelineRunLogger(nil, 2)
	logger3 := NewPipelineRunLogger(nil, 4)

	logger.SetLevel(logrus.InfoLevel)
	logger2.SetLevel(logrus.InfoLevel)
	logger3.SetLevel(logrus.InfoLevel)

	logger.AddReaderEntry(logger2.Reader())
	logger2.AddReaderEntry(logger3.Reader())

	logger.Info(logrus.WithField("message", "logger test entry"))
	logger2.Info(logrus.WithField("message", "logger2 test entry"))
	logger3.Info(logrus.WithField("message", "logger3 test entry"))

	logger3.Close()
	logger2.Close()
	logger.Close()

	result := logger.String()
	require.Contains(t, result, "logger test entry")
	require.Contains(t, result, "logger2 test entry")
	require.Contains(t, result, "logger3 test entry")
}
