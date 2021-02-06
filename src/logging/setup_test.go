package logging

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRoot_SetUpLogs(t *testing.T) {
	log := logrus.New()
	buffer := bytes.NewBuffer([]byte{})

	err := SetUpLogs(log, "", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.WarnLevel, UserPipeLogLevel)
	require.Equal(t, logrus.ErrorLevel, BuiltInPipeLogLevel)

	err = SetUpLogs(log, "trace", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.TraceLevel, UserPipeLogLevel)
	require.Equal(t, logrus.TraceLevel, BuiltInPipeLogLevel)

	err = SetUpLogs(log, "debug", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.DebugLevel, UserPipeLogLevel)
	require.Equal(t, logrus.InfoLevel, BuiltInPipeLogLevel)

	err = SetUpLogs(log, "info", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.InfoLevel, UserPipeLogLevel)
	require.Equal(t, logrus.WarnLevel, BuiltInPipeLogLevel)

	err = SetUpLogs(log, "warn", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.WarnLevel, UserPipeLogLevel)
	require.Equal(t, logrus.ErrorLevel, BuiltInPipeLogLevel)

	err = SetUpLogs(log, "error", buffer)
	require.Nil(t, err)
	require.Equal(t, logrus.ErrorLevel, UserPipeLogLevel)
	require.Equal(t, logrus.ErrorLevel, BuiltInPipeLogLevel)
}

func TestRoot_SetUpLogs_InvalidLevel(t *testing.T) {
	log := logrus.New()
	UserPipeLogLevel = logrus.InfoLevel

	buffer := bytes.NewBuffer([]byte{})
	err := SetUpLogs(log, "bedug", buffer)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not a valid logrus Level")
	require.Equal(t, logrus.InfoLevel, UserPipeLogLevel)
	require.Equal(t, "", buffer.String())
}
