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

// To keep track of log entries while several pipeline runs might be executing asynchronously,
// each run has its own logger. It needs to:
// - implement the io.Reader interface
// - be non-blocking, even when its output is not yet being read
// - keep a running total of the number of log entries at each log level, as well as all errors
// - be nestable in the sense that the output of another logger can be slotted in
// - preserve the order of log entries, including those slotted in
type Logger struct {
	logEntries list.List

	baseLogger  *logrus.Logger
	run         *Run
	Indentation int
	errors      *multierror.Error
	logBuffer   *bytes.Buffer

	logCountTrace   int
	logCountDebug   int
	logCountInfo    int
	logCountWarning int
	logCountError   int

	closureMutex *sync.Mutex
	closed       bool

	completed           bool
	completionMutex     *sync.Mutex
	completionWaitGroup *sync.WaitGroup

	ErrorCallback func(error)

	unreadBuffer []byte
}

func NewPipelineRunLogger(run *Run, indentation int) *Logger {
	logger := logrus.New()
	logger.Formatter = logging.CustomFormatter{}
	logger.SetLevel(logging.UserPipeLogLevel)
	numTraceLogs := 0
	numDebugLogs := 0
	numInfoLogs := 0
	numWarningLogs := 0
	numErrorLogs := 0
	logBuffer := new(bytes.Buffer)
	logger.SetOutput(logBuffer)
	completionWaitGroup := &sync.WaitGroup{}
	completionWaitGroup.Add(1)

	return &Logger{
		baseLogger:  logger,
		run:         run,
		Indentation: indentation,
		errors:      nil,
		logBuffer:   logBuffer,

		logCountTrace:   numTraceLogs,
		logCountDebug:   numDebugLogs,
		logCountInfo:    numInfoLogs,
		logCountWarning: numWarningLogs,
		logCountError:   numErrorLogs,

		completed:           false,
		completionMutex:     &sync.Mutex{},
		completionWaitGroup: completionWaitGroup,
		closureMutex:        &sync.Mutex{},
		closed:              false,

		unreadBuffer: make([]byte, 0, 1024),
	}
}

func NewClosedPipelineRunLoggerWithResult(buffer *bytes.Buffer) *Logger {
	completionWaitGroup := &sync.WaitGroup{}
	return &Logger{
		completed:           true,
		completionWaitGroup: completionWaitGroup,
		closed:              true,
		unreadBuffer:        buffer.Bytes(),
	}
}

func (logger *Logger) Close() {
	defer logger.closureMutex.Unlock()
	logger.closureMutex.Lock()
	if logger.closed {
		return
	}
	logger.closed = true
	logger.completionWaitGroup.Done()
}

func (logger *Logger) Closed() bool {
	return logger.closed
}

func (logger *Logger) Wait() {
	logger.completionWaitGroup.Wait()
}

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

func (logger *Logger) SetLevel(level logrus.Level) {
	logger.baseLogger.SetLevel(level)
}

func (logger *Logger) Level() logrus.Level {
	return logger.baseLogger.Level
}

func (logger *Logger) AddReaderEntry(reader io.Reader) {
	logEntry := fields.EntryWithFields(
		fields.Indentation(logger.Indentation+2),
		fields.WithReader(reader),
	)
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries.PushBack(logEntry)
}

func (logger *Logger) AddWriteCloserEntry() io.WriteCloser {
	pipeReader, pipeWriter := io.Pipe()
	logger.AddReaderEntry(pipeReader)
	return pipeWriter
}

func (logger *Logger) PossibleError(err error) {
	if err == nil {
		return
	}
	logger.Error(err)
}

func (logger *Logger) PossibleErrorWithExplanation(err error, explanation string) {
	if err == nil {
		return
	}
	logger.Error(fmt.Errorf(explanation+" %w", err))
}

func (logger *Logger) StderrOutput(message string, logFields ...fields.LogEntryField) {
	logger.logCountError += 1
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

func (logger *Logger) Error(err error, logFields ...fields.LogEntryField) {
	logger.logCountError += 1
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
		if logger.run.Identifier == nil {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w\n", "anonymous", err))
		} else {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w\n", *logger.run.Identifier, err))
		}
	}
}

func (logger *Logger) WarnWithFields(logFields ...fields.LogEntryField) {
	logger.logCountWarning += 1
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.WarnLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) Warn(entry *logrus.Entry) {
	logger.logCountWarning += 1
	entry.Level = logrus.WarnLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) Info(entry *logrus.Entry) {
	logger.logCountInfo += 1
	entry.Level = logrus.InfoLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) InfoWithFields(logFields ...fields.LogEntryField) {
	logger.logCountInfo += 1
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.InfoLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) Debug(entry *logrus.Entry) {
	logger.logCountDebug += 1
	entry.Level = logrus.DebugLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) DebugWithFields(logFields ...fields.LogEntryField) {
	logger.logCountDebug += 1
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.DebugLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) Trace(entry *logrus.Entry) {
	logger.logCountTrace += 1
	entry.Level = logrus.TraceLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) TraceWithFields(logFields ...fields.LogEntryField) {
	logger.logCountTrace += 1
	logFields = append(logFields, fields.Run(logger.run))
	entry := fields.EntryWithFields(logFields...)
	entry.Level = logrus.TraceLevel
	logger.logEntries.PushBack(entry)
}

func (logger *Logger) TraceCount() int {
	return logger.logCountTrace
}

func (logger *Logger) DebugCount() int {
	return logger.logCountDebug
}
func (logger *Logger) InfoCount() int {
	return logger.logCountInfo
}
func (logger *Logger) WarnCount() int {
	return logger.logCountWarning
}

func (logger *Logger) ErrorCount() int {
	return logger.logCountError
}

func (logger *Logger) Bytes() []byte {
	result, _ := ioutil.ReadAll(logger)
	return result
}

func (logger *Logger) String() string {
	return string(logger.Bytes())
}

func (logger *Logger) LastError() error {
	if logger.errors == nil || logger.errors.Len() == 0 {
		return nil
	}
	return logger.errors.WrappedErrors()[logger.errors.Len()-1]
}

func (logger *Logger) AllErrorMessages() []string {
	result := make([]string, 0, logger.errors.Len())
	for _, err := range logger.errors.WrappedErrors() {
		result = append(result, err.Error())
	}
	return result
}

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
			} else {
				return logger.readFromLogEntry(p, logEntry)
			}
		} else {
			panic("unknown log entry type")
		}
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
	} else {
		return 0, fmt.Errorf("invalid type for `reader` data field")
	}
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
	result, logErr := logging.CustomFormatter{}.Format(indentedLogEntry)
	if logErr != nil {
		logger.Error(logErr)
	}
	if len(result) > len(p) {
		result, logger.unreadBuffer = result[:len(p)], result[len(p):]
	}
	copy(p, result)
	logger.logEntries.Remove(logger.logEntries.Front())
	return len(result), nil
}
