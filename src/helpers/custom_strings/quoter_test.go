package custom_strings

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleQuote(t *testing.T) {
	result := QuoteValue("test '\" string", "single")
	require.Equal(t, "'test \"'\"\" string'", result)
}

func TestDoubleQuote(t *testing.T) {
	result := QuoteValue("test '\" string", "double")
	require.Equal(t, "\"test '\\\" string\"", result)
}

func TestNoQuote(t *testing.T) {
	result := QuoteValue("test '\" string", "none")
	require.Equal(t, "test '\" string", result)
}

func TestDefaultQuote(t *testing.T) {
	result := QuoteValue("test '\" string", "unknown")
	require.Equal(t, "\"test '\\\" string\"", result)
}
