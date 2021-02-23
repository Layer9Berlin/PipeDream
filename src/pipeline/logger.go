package pipeline

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

// Logger keeps track of a pipeline run's log entries while several pipeline runs might be executing asynchronously.
// It needs to:
// - implement the io.Reader interface
// - be non-blocking, even when its output is not yet being read
// - keep a running total of the number of log entries at each log level, as well as all errors
// - be nestable in the sense that the output of another logger can be slotted in
// - preserve the order of log entries, including those slotted in
type Logger struct {
	logEntries *list.List

	baseLogger  *logrus.Logger
	run         *Run
	Indentation int
	errors      *multierror.Error

	logCountTrace   int
	logCountDebug   int
	logCountInfo    int
	logCountWarning int
	logCountError   int

	closureMutex *sync.Mutex
	closed       bool

	ErrorCallback func(error)

	unreadBuffer []byte
}

// NewLogger creates a new Logger
func NewLogger(run *Run, indentation int) *Logger {
	logger := logrus.New()
	logger.Formatter = logging.LogFormatter{}
	logger.SetLevel(logging.UserPipeLogLevel)
	numTraceLogs := 0
	numDebugLogs := 0
	numInfoLogs := 0
	numWarningLogs := 0
	numErrorLogs := 0

	return &Logger{
		baseLogger:  logger,
		run:         run,
		Indentation: indentation,
		errors:      nil,

		logEntries:      list.New(),
		logCountTrace:   numTraceLogs,
		logCountDebug:   numDebugLogs,
		logCountInfo:    numInfoLogs,
		logCountWarning: numWarningLogs,
		logCountError:   numErrorLogs,

		closureMutex: &sync.Mutex{},
		closed:       false,

		unreadBuffer: make([]byte, 0, 1024),
	}
}

// NewClosedLoggerWithResult creates a new Logger that is already closed with the specified result
func NewClosedLoggerWithResult(buffer *bytes.Buffer) *Logger {
	return &Logger{
		closed:       true,
		unreadBuffer: buffer.Bytes(),
	}
}

// Close finalizes the log, preventing further entries from being logged
func (logger *Logger) Close() {
	defer logger.closureMutex.Unlock()
	logger.closureMutex.Lock()
	if logger.closed {
		return
	}
	logger.closed = true
}

// Closed indicates whether the run has been closed
func (logger *Logger) Closed() bool {
	return logger.closed
}

// Summary returns a human-readable short description of the logged warnings and errors
func (logger *Logger) Summary() string {
	components := make([]string, 0, 2)
	if logger.logCountWarning > 0 {
		components = append(components, fmt.Sprint(aurora.Yellow(fmt.Sprint("⚠️", logger.logCountWarning))))
	}
	if logger.logCountError > 0 {
		components = append(components, fmt.Sprint(aurora.Red(fmt.Sprint("🛑", logger.logCountError))))
	}
	return strings.Join(components, "  ")
}

// SetLevel sets the logger's log level
func (logger *Logger) SetLevel(level logrus.Level) {
	logger.baseLogger.SetLevel(level)
}

// Level indicates the log level
func (logger *Logger) Level() logrus.Level {
	return logger.baseLogger.Level
}

// AddReaderEntry adds an entry that will write the entire contents of the provided reader before proceeding to the next entry
func (logger *Logger) AddReaderEntry(reader io.Reader) {
	logEntry := fields.EntryWithFields(
		fields.Indentation(logger.Indentation+2),
		fields.WithReader(reader),
	)
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries.PushBack(logEntry)
}

// AddWriteCloserEntry adds an entry that will write everything written to the writer to the log before proceeding to the next entry
//
// You must close the io.WriteCloser to indicate that you are done writing, so that the log can move on.
func (logger *Logger) AddWriteCloserEntry() io.WriteCloser {
	pipeReader, pipeWriter := io.Pipe()
	logger.AddReaderEntry(pipeReader)
	return pipeWriter
}

// PossibleError does nothing if it is passed nil and otherwise logs the provided error
func (logger *Logger) PossibleError(err error) {
	if err == nil {
		return
	}
	logger.Error(err)
}

// PossibleErrorWithExplanation does nothing if it is passed nil and otherwise logs the provided error with an additional explanation
func (logger *Logger) PossibleErrorWithExplanation(err error, explanation string) {
	if err == nil {
		return
	}
	logger.Error(fmt.Errorf(explanation+" %w", err))
}

// StderrOutput adds an appropriate entry for non-trivial stderr output to the logs
func (logger *Logger) StderrOutput(message string, logFields ...fields.LogEntryField) {
	logger.logCountError++
	logger.errors = multierror.Append(logger.errors, fmt.Errorf("stderr: %v", message))
	logEntry := logrus.WithFields(logrus.Fields{
		"prefix":  "⛔️ ",
		"message": message,
	})
	logFields = append(logFields, fields.Run(logger.run))
	for _, withField := range logFields {
		logEntry = withField(logEntry)
	}
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries.PushBack(logEntry)
}

// Error adds an appropriate entry for an encountered error
func (logger *Logger) Error(err error, logFields ...fields.LogEntryField) {
	logger.logCountError++
	logger.errors = multierror.Append(logger.errors, err)
	logEntry := logrus.WithFields(logrus.Fields{
		"prefix":  "🛑 ",
		"message": err.Error(),
	})
	logFields = append(logFields, fields.Run(logger.run))
	for _, withField := range logFields {
		logEntry = withField(logEntry)
	}
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries.PushBack(logEntry)
	if logger.ErrorCallback != nil {
		if logger.run == nil || logger.run.Identifier == nil {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w", "anonymous", err))
		} else {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w", *logger.run.Identifier, err))
		}
	}
}

// Warn adds an appropriate entry for an encountered warning
func (logger *Logger) Warn(logFields ...fields.LogEntryField) {
	logger.logCountWarning++
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.WarnLevel
	logger.logEntries.PushBack(entry)
}

// Info adds an appropriate entry for non-critical information
func (logger *Logger) Info(logFields ...fields.LogEntryField) {
	logger.logCountInfo++
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.InfoLevel
	logger.logEntries.PushBack(entry)
}

// Debug adds an appropriate entry for a debug message
func (logger *Logger) Debug(logFields ...fields.LogEntryField) {
	logger.logCountDebug++
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.DebugLevel
	logger.logEntries.PushBack(entry)
}

// Trace adds an appropriate entry for a trace message
func (logger *Logger) Trace(logFields ...fields.LogEntryField) {
	logger.logCountTrace++
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.TraceLevel
	logger.logEntries.PushBack(entry)
}

// TraceCount is the total number of trace logs encountered
func (logger *Logger) TraceCount() int {
	return logger.logCountTrace
}

// DebugCount is the total number of debug logs encountered
func (logger *Logger) DebugCount() int {
	return logger.logCountDebug
}

// InfoCount is the total number of info logs encountered
func (logger *Logger) InfoCount() int {
	return logger.logCountInfo
}

// WarnCount is the total number of warning logs encountered
func (logger *Logger) WarnCount() int {
	return logger.logCountWarning
}

// ErrorCount is the total number of error logs encountered
func (logger *Logger) ErrorCount() int {
	return logger.logCountError
}

// String returns the logger's total output as a string
//
// Note that this will consume the output, competing with other reads,
// so String should only be called once and not in conjunction with Read.
func (logger *Logger) String() string {
	result, _ := ioutil.ReadAll(logger)
	return string(result)
}

// LastError returns the most recent error level log entry
func (logger *Logger) LastError() error {
	if logger.errors == nil || logger.errors.Len() == 0 {
		return nil
	}
	return logger.errors.WrappedErrors()[logger.errors.Len()-1]
}

// AllErrorMessages returns all errors logged up to this point
func (logger *Logger) AllErrorMessages() []string {
	result := make([]string, 0, logger.errors.Len())
	for _, err := range logger.errors.WrappedErrors() {
		result = append(result, err.Error())
	}
	return result
}

// Read writes an entry, or part of an entry to the provided buffer
//
// io.EOF indicates that the log is closed and all log entries have been read.
func (logger *Logger) Read(p []byte) (int, error) {
	if logger.unreadBuffer != nil && len(logger.unreadBuffer) > 0 {
		return logger.readFromUnreadBuffer(p)
	}
	firstItem := logger.logEntries.Front()
	if firstItem != nil {
		if logEntry, ok := firstItem.Value.(*logrus.Entry); ok {
			readerDataEntry := logEntry.Data["reader"]
			if readerDataEntry != nil {
				return logger.readFromNestedReader(p, readerDataEntry)
			}
			return logger.readFromLogEntry(p, logEntry)
		}
		panic(fmt.Sprintf("unknown log entry type: %T", firstItem.Value))
	}
	if logger.Closed() {
		return 0, io.EOF
	}
	return 0, nil
}

func (logger *Logger) readFromUnreadBuffer(p []byte) (int, error) {
	result := logger.unreadBuffer
	if len(logger.unreadBuffer) > len(p) {
		result, logger.unreadBuffer = result[:len(p)], result[len(p):]
	} else {
		logger.unreadBuffer = make([]byte, 0)
	}
	copy(p, result)
	return len(result), nil
}

func (logger *Logger) readFromNestedReader(p []byte, readerDataEntry interface{}) (int, error) {
	if reader, ok := readerDataEntry.(io.Reader); ok {
		n, err := reader.Read(p)
		if err == io.EOF {
			logger.logEntries.Remove(logger.logEntries.Front())
			err = nil
		}
		return n, err
	}

	return 0, fmt.Errorf("invalid type for `reader` data field")
}

func (logger *Logger) readFromLogEntry(p []byte, logEntry *logrus.Entry) (int, error) {
	// discard entries whose log level is above (i.e. of lower severity) than the logger's level
	if logEntry.Level > logger.Level() {
		logger.logEntries.Remove(logger.logEntries.Front())
		return 0, nil
	}
	logEntry.Logger = logger.baseLogger
	indentation := logger.Indentation
	if logger.baseLogger.Level <= logrus.InfoLevel {
		// we don't indent at this log level, as there are too few messages to make it worthwhile
		indentation = 0
	}
	indentedLogEntry := logEntry.WithField("indentation", indentation)
	indentedLogEntry.Level = logEntry.Level
	result, _ := logging.LogFormatter{}.Format(indentedLogEntry)
	if len(result) > len(p) {
		result, logger.unreadBuffer = result[:len(p)], result[len(p):]
	}
	copy(p, result)
	logger.logEntries.Remove(logger.logEntries.Front())
	return len(result), nil
}
