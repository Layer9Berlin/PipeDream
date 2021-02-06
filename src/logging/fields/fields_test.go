package fields

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/logrusorgru/aurora/v3"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestLogFields_Message(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(Message("test message")),
	)
	require.Contains(t, string(result), "test message")
}

func TestLogFields_Info(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(Info("test info")),
	)
	require.Contains(t, string(result), "test info")
}

func TestLogFields_Symbol(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(Symbol("😇")),
	)
	require.Contains(t, string(result), "😇")
}

func TestLogFields_Color(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(
			Message("test message"),
			Color("green"),
		),
	)
	require.Equal(t, fmt.Sprint(aurora.Green("test message"), "\n"), string(result))
}

func TestLogFields_Middleware(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(
			Middleware("test middleware"),
		),
	)
	require.Equal(t, fmt.Sprint("test middleware", "\n"), string(result))
}

func TestLogFields_Indentation(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(
			Indentation(0),
		),
	)
	require.NotContains(t, string(result), "    ")
	result, _ = logging.CustomFormatter{}.Format(
		EntryWithFields(
			Indentation(4),
		),
	)
	require.Contains(t, string(result), "    ")
}

func TestMiddlewareLogEntries_WithReader(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(WithReader(strings.NewReader("test"))),
	)
	require.Equal(t, "test", string(result))
}

func TestMiddlewareLogEntries_DataStream(t *testing.T) {
	result, _ := logging.CustomFormatter{}.Format(
		EntryWithFields(DataStream("test", "message")...),
	)
	require.Equal(t, fmt.Sprint(aurora.Gray(18, "⎇ test | message"), "\n"), string(result))
}
