package log_fields

import (
	"github.com/sirupsen/logrus"
	"io"
)

type LogEntryField func(*logrus.Entry) *logrus.Entry

func Message(message interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("message", message)
	}
}

func Info(info interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("info", info)
	}
}

func Symbol(symbol string) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("prefix", symbol+" ")
	}
}

func Color(color string) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("color", color)
	}
}

func Run(run interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("run", run)
	}
}

func Middleware(middleware interface{}) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("middleware", middleware)
	}
}

func Indentation(indentation int) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("indentation", indentation)
	}
}

func WithReader(reader io.Reader) LogEntryField {
	return func(entry *logrus.Entry) *logrus.Entry {
		return entry.WithField("reader", reader)
	}
}

func DataStream(middleware interface{}, message string) []LogEntryField {
	return []LogEntryField{
		Symbol("⎇"),
		Middleware(middleware),
		Message(message),
		Color("lightgray"),
	}
}

func EntryWithFields(fields ...LogEntryField) *logrus.Entry {
	entry := logrus.WithFields(logrus.Fields{})
	for _, withField := range fields {
		entry = withField(entry)
	}
	return entry
}
