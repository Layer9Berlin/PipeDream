package logging

import (
	"github.com/sirupsen/logrus"
	"io"
)

func SetUpLogs(log *logrus.Logger, verbosity string, out io.Writer) error {
	log.SetOutput(out)
	if verbosity == "" {
		verbosity = logrus.WarnLevel.String()
	}
	logLevel, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	log.SetLevel(logLevel)
	// let's be nice and log built-in pipes and parsers
	// with less detail than user defined pipes
	UserPipeLogLevel = logLevel
	internalLogLevel := logLevel
	switch logLevel {
	case logrus.ErrorLevel:
		internalLogLevel = logrus.ErrorLevel
	case logrus.WarnLevel:
		internalLogLevel = logrus.ErrorLevel
	case logrus.InfoLevel:
		internalLogLevel = logrus.WarnLevel
	case logrus.DebugLevel:
		internalLogLevel = logrus.InfoLevel
	default:
	}
	BuiltInPipeLogLevel = internalLogLevel
	log.Tracef("log level set to %q for user-defined pipes and %q for built-in pipes", logLevel, internalLogLevel)
	return nil
}
