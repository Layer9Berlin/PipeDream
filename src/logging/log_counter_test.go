package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLogCounterHook_Fire(t *testing.T) {
	counter := 8
	hook := NewLogCounterHook(logrus.ErrorLevel, &counter)
	err := hook.Fire(logrus.WithError(fmt.Errorf("error")))
	require.Nil(t, err)
	require.Equal(t, 9, counter)
}

func TestLogCounterHook_Levels(t *testing.T) {
	counter := 8
	hook := NewLogCounterHook(logrus.ErrorLevel, &counter)
	require.Equal(t, []logrus.Level{logrus.ErrorLevel}, hook.Levels())
}

