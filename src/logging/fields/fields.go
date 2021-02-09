// Package fields provides custom log entry fields
package fields

import (
	"github.com/sirupsen/logrus"
	"io"
)

// LogEntryField is a convenience function wrapper for log entry fields
//
// Apply the function in order to make the required changes to a log entry
type LogEntryField func(*logrus.Entry) *logrus.Entry

// Message adds a message to the log entry
func Message(message interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("message", message)
	}
}

// Info adds additional information to the log entry
func Info(info interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("info", info)
	}
}

// Symbol adds a symbol to illustrate the log entry
func Symbol(symbol string) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("prefix", symbol+" ")
	}
}

// Color provides a way of coloring a log entry
//
// If no color is provided, a color will be chosen based on the log level.
func Color(color string) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("color", color)
	}
}

// Run adds information about the pipeline run context to the log entry
func Run(run interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("run", run)
	}
}

// Middleware adds information about the middleware context to the log entry
func Middleware(middleware interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("middleware", middleware)
	}
}

// Indentation sets an indentation level for the log entry
//
// Note that indentation may be ignored at some log levels.
func Indentation(indentation int) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("indentation", indentation)
	}
}

// WithReader adds an io.Reader to the log entry
//
// The data provided by the reader will be output to the log
// before the next entry. This is useful for nesting logs.
func WithReader(reader io.Reader) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("reader", reader)
	}
}

// DataStream is a convenience function for data stream logs
func DataStream(middleware interface{}, message string) []LogEntryField {
	return []LogEntryField{
		Symbol("⎇"),
		Middleware(middleware),
		Message(message),
		Color("lightgray"),
	}
}

// EntryWithFields returns a new log entry containing the provided fields
func EntryWithFields(fields ...LogEntryField) *logrus.Entry {
	entry := logrus.WithFields(logrus.Fields{})
	for _, withField := range fields {
		entry = withField(entry)
	}
	return entry
}
