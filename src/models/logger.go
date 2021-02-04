package models

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"io"
	"pipedream/src/logging"
	"pipedream/src/logging/log_fields"
	"strings"
	"sync"
)

type PipelineRunLogger struct {
	logEntries []*logrus.Entry

	baseLogger  *logrus.Logger
	run         *PipelineRun
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
}

func NewPipelineRunLogger(run *PipelineRun, indentation int) *PipelineRunLogger {
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

	return &PipelineRunLogger{
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
	}
}

func NewClosedPipelineRunLoggerWithResult(buffer *bytes.Buffer) *PipelineRunLogger {
	completionWaitGroup := &sync.WaitGroup{}
	return &PipelineRunLogger{
		completed:           true,
		completionWaitGroup: completionWaitGroup,
		closed:              true,
		logBuffer:           buffer,
	}
}

func NewClosedPipelineRunLoggerWithErrors(errors *multierror.Error) *PipelineRunLogger {
	completionWaitGroup := &sync.WaitGroup{}
	return &PipelineRunLogger{
		completed:           true,
		completionWaitGroup: completionWaitGroup,
		closed:              true,
		logCountError:       errors.Len(),
		errors:              errors,
	}
}

func (logger *PipelineRunLogger) Close() {
	logger.closureMutex.Lock()
	defer logger.closureMutex.Unlock()
	if logger.closed {
		return
	}
	logger.closed = true

	// we don't expect any more inputs
	go func() {
		defer logger.completionWaitGroup.Done()
		for _, logEntry := range logger.logEntries {
			logEntry.Logger = logger.baseLogger
			indentation := logger.Indentation
			//if indentationField, haveIndentationField := logEntry.Data["indentation"]; haveIndentationField {
			//	if indentationAsInt, indentationIsInt := indentationField.(int); indentationIsInt {
			//		indentation = indentationAsInt
			//	}
			//}
			if logger.baseLogger.Level <= logrus.InfoLevel {
				// we don't indent at this log level, as there are too few messages to make it worthwhile
				indentation = 0
			}
			logEntry.WithField("indentation", indentation).Log(logEntry.Level)
		}
		logger.completed = true
	}()
}

func (logger *PipelineRunLogger) Closed() bool {
	return logger.closed
}

func (logger *PipelineRunLogger) Wait() {
	logger.completionWaitGroup.Wait()
}

func (logger *PipelineRunLogger) Completed() bool {
	return logger.completed
}

func (logger *PipelineRunLogger) Summary() string {
	components := make([]string, 0, 2)
	if logger.logCountWarning > 0 {
		components = append(components, fmt.Sprint(aurora.Yellow(fmt.Sprint("⚠️", logger.logCountWarning))))
	}
	if logger.logCountError > 0 {
		components = append(components, fmt.Sprint(aurora.Red(fmt.Sprint("🛑", logger.logCountError))))
	}
	return strings.Join(components, "  ")
}

func (logger *PipelineRunLogger) SetLevel(level logrus.Level) {
	logger.baseLogger.SetLevel(level)
}

func (logger *PipelineRunLogger) Level() logrus.Level {
	return logger.baseLogger.Level
}

func (logger *PipelineRunLogger) AddReaderEntry(reader io.Reader) {
	logEntry := log_fields.EntryWithFields(
		log_fields.Indentation(logger.Indentation+2),
		log_fields.WithReader(reader),
	)
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries = append(logger.logEntries, logEntry)
}

func (logger *PipelineRunLogger) AddWriteCloserEntry() io.WriteCloser {
	pipeReader, pipeWriter := io.Pipe()
	logger.AddReaderEntry(pipeReader)
	return pipeWriter
}

func (logger *PipelineRunLogger) PossibleError(err error) {
	if err == nil {
		return
	}
	logger.Error(err)
}

func (logger *PipelineRunLogger) PossibleErrorWithExplanation(err error, explanation string) {
	if err == nil {
		return
	}
	logger.Error(fmt.Errorf(explanation+" %w", err))
}

func (logger *PipelineRunLogger) StderrOutput(message string, fields ...log_fields.LogEntryField) {
	logger.logCountError += 1
	logger.errors = multierror.Append(logger.errors, fmt.Errorf("stderr: %v", message))
	logEntry := logrus.WithFields(logrus.Fields{
		"prefix":  "⛔️ ",
		"message": message,
	})
	fields = append(fields, log_fields.Run(logger.run))
	for _, withField := range fields {
		logEntry = withField(logEntry)
	}
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries = append(logger.logEntries, logEntry)
}

func (logger *PipelineRunLogger) Error(err error, fields ...log_fields.LogEntryField) {
	logger.logCountError += 1
	logger.errors = multierror.Append(logger.errors, err)
	logEntry := logrus.WithFields(logrus.Fields{
		"prefix":  "🛑 ",
		"message": err.Error(),
	})
	fields = append(fields, log_fields.Run(logger.run))
	for _, withField := range fields {
		logEntry = withField(logEntry)
	}
	logEntry.Level = logrus.ErrorLevel
	logger.logEntries = append(logger.logEntries, logEntry)
	if logger.ErrorCallback != nil {
		if logger.run.Identifier == nil {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w\n", "anonymous", err))
		} else {
			logger.ErrorCallback(fmt.Errorf("%v:\n%w\n", *logger.run.Identifier, err))
		}
	}
}

func (logger *PipelineRunLogger) WarnWithFields(fields ...log_fields.LogEntryField) {
	logger.logCountWarning += 1
	fields = append(fields, log_fields.Run(logger.run))
	entry := log_fields.EntryWithFields(fields...)
	entry.Level = logrus.WarnLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) Warn(entry *logrus.Entry) {
	logger.logCountWarning += 1
	entry.Level = logrus.WarnLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) Info(entry *logrus.Entry) {
	logger.logCountInfo += 1
	entry.Level = logrus.InfoLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) InfoWithFields(fields ...log_fields.LogEntryField) {
	logger.logCountInfo += 1
	fields = append(fields, log_fields.Run(logger.run))
	entry := log_fields.EntryWithFields(fields...)
	entry.Level = logrus.InfoLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) Debug(entry *logrus.Entry) {
	logger.logCountDebug += 1
	entry.Level = logrus.DebugLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) DebugWithFields(fields ...log_fields.LogEntryField) {
	logger.logCountDebug += 1
	fields = append(fields, log_fields.Run(logger.run))
	entry := log_fields.EntryWithFields(fields...)
	entry.Level = logrus.DebugLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) Trace(entry *logrus.Entry) {
	logger.logCountTrace += 1
	entry.Level = logrus.TraceLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) TraceWithFields(fields ...log_fields.LogEntryField) {
	logger.logCountTrace += 1
	fields = append(fields, log_fields.Run(logger.run))
	entry := log_fields.EntryWithFields(fields...)
	entry.Level = logrus.TraceLevel
	logger.logEntries = append(logger.logEntries, entry)
}

func (logger *PipelineRunLogger) TraceCount() int {
	return logger.logCountTrace
}

func (logger *PipelineRunLogger) DebugCount() int {
	return logger.logCountDebug
}
func (logger *PipelineRunLogger) InfoCount() int {
	return logger.logCountInfo
}
func (logger *PipelineRunLogger) WarnCount() int {
	return logger.logCountWarning
}

func (logger *PipelineRunLogger) ErrorCount() int {
	return logger.logCountError
}

func (logger *PipelineRunLogger) Bytes() []byte {
	logger.Wait()
	return logger.logBuffer.Bytes()
}

func (logger *PipelineRunLogger) String() string {
	logger.Wait()
	return logger.logBuffer.String()
}

func (logger *PipelineRunLogger) LastError() error {
	if logger.errors == nil || logger.errors.Len() == 0 {
		return nil
	}
	return logger.errors.WrappedErrors()[logger.errors.Len()-1]
}

func (logger *PipelineRunLogger) AllErrorMessages() []string {
	result := make([]string, 0, logger.errors.Len())
	for _, err := range logger.errors.WrappedErrors() {
		result = append(result, err.Error())
	}
	return result
}

func (logger *PipelineRunLogger) Reader() io.Reader {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		logger.Wait()
		_, err := pipeWriter.Write(logger.Bytes())
		if err != nil {
			panic(err)
		}
		err = pipeWriter.Close()
		if err != nil {
			panic(err)
		}
	}()
	return pipeReader
}
