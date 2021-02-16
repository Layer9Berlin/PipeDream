package pipeline

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestPipelineRunLogger_PossibleError(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.PossibleError(nil)
	logger.PossibleError(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
}

func TestPipelineRunLogger_Error(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.Error(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
	require.Equal(t, "test error", logger.LastError().Error())
	logger.Error(fmt.Errorf("another error"), fields.Middleware("test 2"))
	require.Equal(t, 2, logger.ErrorCount())
	require.Equal(t, "another error", logger.LastError().Error())
}

func TestPipelineRunLogger_Counts(t *testing.T) {
	logger := NewLogger(nil, 0)

	require.Equal(t, 0, logger.TraceCount())
	logger.Trace()
	require.Equal(t, 1, logger.TraceCount())

	require.Equal(t, 0, logger.DebugCount())
	logger.Debug()
	require.Equal(t, 1, logger.DebugCount())

	require.Equal(t, 0, logger.InfoCount())
	logger.Info()
	require.Equal(t, 1, logger.InfoCount())

	require.Equal(t, 0, logger.WarnCount())
	logger.Warn()
	require.Equal(t, 1, logger.WarnCount())

	require.Equal(t, 0, logger.ErrorCount())
	logger.Error(fmt.Errorf("test error"))
	require.Equal(t, 1, logger.ErrorCount())
}

func TestPipelineRunLogger_Close(t *testing.T) {
	logger := NewLogger(nil, 0)

	require.False(t, logger.Closed())

	logger.Close()

	require.True(t, logger.Closed())
}

func TestPipelineRunLogger_PossibleErrorWithExplanation(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.PossibleErrorWithExplanation(nil, "just testing: ")
	require.Equal(t, 0, logger.ErrorCount())

	logger.PossibleErrorWithExplanation(fmt.Errorf("test error"), "just testing:")
	require.Equal(t, 1, logger.ErrorCount())
	require.Equal(t, "just testing: test error", logger.LastError().Error())
}

func TestPipelineRunLogger_LastError(t *testing.T) {
	logger := NewLogger(nil, 0)

	require.Nil(t, logger.LastError())
	err1 := fmt.Errorf("test error 1")
	logger.Error(err1)
	require.Equal(t, err1, logger.LastError())
	err2 := fmt.Errorf("test error 2")
	logger.Error(err2)
	require.Equal(t, err2, logger.LastError())
}

func TestPipelineRunLogger_AllErrorMessages(t *testing.T) {
	logger := NewLogger(nil, 0)

	err1 := fmt.Errorf("test error 1")
	logger.Error(err1)
	err2 := fmt.Errorf("test error 2")
	logger.Error(err2)
	require.Equal(t, []string{"test error 1", "test error 2"}, logger.AllErrorMessages())
}

func TestPipelineRunLogger_CloseAlreadyClosed(t *testing.T) {
	logger := NewLogger(nil, 0)

	// closing multiple times is supported
	logger.Close()
	logger.Close()
	logger.Close()

	require.True(t, logger.Closed())
}

func TestPipelineRunLogger_Output(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.SetLevel(logrus.TraceLevel)
	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.Close()

	logString := logger.String()
	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	), logString)
	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	), logString)
}

func TestPipelineRunLogger_WriteToReader_TraceLevel(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.SetLevel(logrus.TraceLevel)
	logger.Close()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
		aurora.Gray(18, "test trace"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_DebugLevel(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.SetLevel(logrus.DebugLevel)
	logger.Close()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
		aurora.Gray(12, "test debug"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_InfoLevel(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.SetLevel(logrus.InfoLevel)
	logger.Close()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
		aurora.Blue("test info"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_WarningLevel(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.SetLevel(logrus.WarnLevel)
	logger.Close()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
		aurora.Yellow("test warning"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_WriteToReader_ErrorLevel(t *testing.T) {
	logger := NewLogger(nil, 0)

	logger.Error(fmt.Errorf("test error"))
	logger.Warn(fields.Message("test warning"))
	logger.Info(fields.Message("test info"))
	logger.Debug(fields.Message("test debug"))
	logger.Trace(fields.Message("test trace"))

	logger.SetLevel(logrus.ErrorLevel)
	logger.Close()

	require.Equal(t, fmt.Sprint(
		aurora.Red("🛑 test error"), "\n",
	), logger.String())
}

func TestPipelineRunLogger_NestedReaders(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger2 := NewLogger(nil, 2)
	logger3 := NewLogger(nil, 4)

	logger.SetLevel(logrus.InfoLevel)
	logger2.SetLevel(logrus.InfoLevel)
	logger3.SetLevel(logrus.InfoLevel)

	logger.AddReaderEntry(logger2)
	logger2.AddReaderEntry(logger3)

	logger.Info(fields.Message("logger test entry"))
	logger2.Info(fields.Message("logger2 test entry"))
	logger3.Info(fields.Message("logger3 test entry"))

	logger3.Close()
	logger2.Close()
	logger.Close()

	result := logger.String()
	require.Contains(t, result, "logger test entry")
	require.Contains(t, result, "logger2 test entry")
	require.Contains(t, result, "logger3 test entry")
}

func TestLogger_Read_unreadBuffer(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.unreadBuffer = []byte("test")
	buffer := make([]byte, 1024)
	count, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, 4, count)
	require.Equal(t, "test", string(buffer[:4]))
}

func TestLogger_Read_largeUnreadBuffer(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.unreadBuffer = []byte("test")
	buffer := make([]byte, 2)
	count, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, 2, count)
	require.Equal(t, "te", string(buffer[:2]))
}

func TestLogger_Read_empty(t *testing.T) {
	logger := NewLogger(nil, 0)
	buffer := make([]byte, 1024)
	count, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, 0, count)
}

func TestLogger_Read_unknownLogEntryType(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic did not occur")
		}
		require.Equal(t, "unknown log entry type: string", r)
	}()
	logger := NewLogger(nil, 0)
	logger.logEntries.PushBack("test")
	buffer := make([]byte, 1024)
	_, err := logger.Read(buffer)
	require.NotNil(t, err)
}

func TestLogger_readFromLogEntry_longEntry(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.Info(fields.Message("test"))
	buffer := make([]byte, 2)
	count, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, 2, count)
}

func TestLogger_NewClosedLoggerWithResult(t *testing.T) {
	logger := NewClosedLoggerWithResult(bytes.NewBuffer([]byte("test")))
	require.Equal(t, "test", logger.String())
}

func TestLogger_readFromNestedReader_invalidType(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.logEntries.PushBack(logrus.WithFields(logrus.Fields{"reader": "invalid"}))
	buffer := make([]byte, 1024)
	count, err := logger.Read(buffer)
	require.NotNil(t, err)
	require.Equal(t, "invalid type for `reader` data field", err.Error())
	require.Equal(t, 0, count)
}

func TestLogger_StderrOutput(t *testing.T) {
	logger := NewLogger(nil, 0)
	logger.StderrOutput("test stderr output")
	buffer := make([]byte, 1024)
	count, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, 35, count)
	require.Contains(t, string(buffer), "test stderr output")

}

func TestLogger_Error_WithCallback_anonymous(t *testing.T) {
	logger := NewLogger(nil, 0)
	errMessage := ""
	logger.ErrorCallback = func(err error) {
		errMessage = err.Error()
	}
	logger.Error(fmt.Errorf("test error"))
	buffer := make([]byte, 1024)
	_, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, "anonymous:\ntest error", errMessage)
}

func TestLogger_Error_WithCallback_namedPipe(t *testing.T) {
	identifier := "test"
	run, _ := NewRun(&identifier, nil, nil, nil)
	logger := NewLogger(run, 0)
	errMessage := ""
	logger.ErrorCallback = func(err error) {
		errMessage = err.Error()
	}
	logger.Error(fmt.Errorf("test error"))
	buffer := make([]byte, 1024)
	_, err := logger.Read(buffer)
	require.Nil(t, err)
	require.Equal(t, "test:\ntest error", errMessage)
}

func TestLogger_AddWriteCloserEntry(t *testing.T) {
	logger := NewLogger(nil, 0)
	writeCloser := logger.AddWriteCloserEntry()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		_, _ = writeCloser.Write([]byte("test"))
		_ = writeCloser.Close()
	}()
	result := ""

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		buffer := make([]byte, 1024)
		count, err := logger.Read(buffer)
		require.Nil(t, err)
		result = string(buffer[:count])
	}()
	waitGroup.Wait()

	require.Equal(t, "test", result)
}
