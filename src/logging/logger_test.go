package logging

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestLogger_FormatWithReader(t *testing.T) {
	reader := strings.NewReader("test")
	logger := logrus.New()
	result, err := CustomFormatter{}.Format(fields.WithReader(reader)(logrus.NewEntry(logger)))
	require.Nil(t, err)
	require.Equal(t, "test", string(result))
}

func TestLogger_Indentation(t *testing.T) {
	logger := logrus.New()
	result, err := CustomFormatter{}.Format(logrus.NewEntry(logger).WithField("info", "test"))
	require.Nil(t, err)
	require.Equal(t, "test\n", string(result))

	result, err = CustomFormatter{}.Format(logrus.NewEntry(logger).WithField("info", "test").WithField("indentation", 4))
	require.Nil(t, err)
	require.Equal(t, "    test\n", string(result))
}

func TestLogger_EmptyMapInfo(t *testing.T) {
	logger := logrus.New()
	result, err := CustomFormatter{}.Format(logrus.NewEntry(logger).WithField("info", make(map[string]interface{}, 0)))
	require.Nil(t, err)
	require.Equal(t, "\n", string(result))
}

func TestLogger_Map(t *testing.T) {
	logger := logrus.New()
	result, err := CustomFormatter{}.Format(logrus.NewEntry(logger).WithField("info", map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}))
	require.Nil(t, err)
	require.Equal(t, "{ key1: `value1`, key2: `value2` }\n", string(result))
}

func TestLogger_ShortenString(t *testing.T) {
	require.Equal(t, "extrem…", ShortenString("extremely long string 1234732 237482 347657943543534657 3240524523", 6))
	require.Equal(t, "test1⇤\xef\xb8…", ShortenString("test1\r\ntest2\r\n", 10))
}

func TestLogger_MessageAndInfo(t *testing.T) {
	logger := logrus.New()
	log, err := CustomFormatter{}.Format(logrus.NewEntry(logger).
		WithField("message", "some message").
		WithField("info", "info as string"))
	require.Nil(t, err)
	require.Contains(t, string(log), "some message | info as string")
}

func TestLogger_NoSeparator(t *testing.T) {
	logger := logrus.New()
	log, err := CustomFormatter{}.Format(logrus.NewEntry(logger).
		WithField("message", "some message"))
	require.Nil(t, err)
	require.NotContains(t, string(log), "|")
}

func TestLogger_Array(t *testing.T) {
	logger := logrus.New()
	log, err := CustomFormatter{}.Format(logrus.NewEntry(logger).
		WithField("message", []string{"some message", "another message"}))
	require.Nil(t, err)
	require.Contains(t, string(log), "some message, another message")
}

func TestLogger_Colors(t *testing.T) {
	logger := logrus.New()
	entry := logrus.NewEntry(logger).
		WithField("message", "some message")
	entry.Level = logrus.ErrorLevel
	log, _ := CustomFormatter{}.Format(entry)
	// red
	require.Contains(t, string(log), "\x1b[31m")

	entry.Level = logrus.WarnLevel
	log, _ = CustomFormatter{}.Format(entry)
	// amber
	require.Contains(t, string(log), "\x1b[33m")

	entry.Level = logrus.InfoLevel
	log, _ = CustomFormatter{}.Format(entry)
	// blue
	require.Contains(t, string(log), "\x1b[34m")

	entry.Level = logrus.DebugLevel
	log, _ = CustomFormatter{}.Format(entry)
	// grey
	require.Contains(t, string(log), "\x1b[38;5;244m")

	entry.Level = logrus.TraceLevel
	log, _ = CustomFormatter{}.Format(entry)
	// light grey
	require.Contains(t, string(log), "\x1b[38;5;250m")
}

func TestLogger_ColorOverrides(t *testing.T) {
	entry := fields.EntryWithFields(
		fields.Message("some message"),
		fields.Color("red"),
	)
	log, _ := CustomFormatter{}.Format(entry)
	// red
	require.Contains(t, string(log), "\x1b[31m")

	entry = fields.Color("yellow")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// amber
	require.Contains(t, string(log), "\x1b[33m")

	entry = fields.Color("blue")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// blue
	require.Contains(t, string(log), "\x1b[34m")

	entry = fields.Color("gray")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// gray
	require.Contains(t, string(log), "\x1b[38;5;244m")

	entry = fields.Color("lightgray")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// light gray
	require.Contains(t, string(log), "\x1b[38;5;250m")

	entry = fields.Color("cyan")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// cyan
	require.Contains(t, string(log), "\x1b[36m")

	entry = fields.Color("black")(entry)
	log, _ = CustomFormatter{}.Format(entry)
	// black
	require.Contains(t, string(log), "\x1b[30m")
}
