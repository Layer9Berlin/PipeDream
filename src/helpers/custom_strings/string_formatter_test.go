package custom_strings

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrettyPrintedByteCount_Bytes(t *testing.T) {
	require.Equal(t, "0B", PrettyPrintedByteCount(0))
	require.Equal(t, "4B", PrettyPrintedByteCount(4))
	require.Equal(t, "12B", PrettyPrintedByteCount(12))
	require.Equal(t, "99B", PrettyPrintedByteCount(99))
	require.Equal(t, "999B", PrettyPrintedByteCount(999))
}

func TestPrettyPrintedByteCount_Kilobytes(t *testing.T) {
	require.Equal(t, "1.0kB", PrettyPrintedByteCount(1_000))
	require.Equal(t, "1.2kB", PrettyPrintedByteCount(1_234))
	require.Equal(t, "500.0kB", PrettyPrintedByteCount(500_000))
	require.Equal(t, "1000.0kB", PrettyPrintedByteCount(999_999))
}

func TestPrettyPrintedByteCount_Megabytes(t *testing.T) {
	require.Equal(t, "1.0MB", PrettyPrintedByteCount(1_000_000))
	require.Equal(t, "1.2MB", PrettyPrintedByteCount(1_234_567))
	require.Equal(t, "500.0MB", PrettyPrintedByteCount(500_000_000))
	require.Equal(t, "1000.0MB", PrettyPrintedByteCount(999_999_999))
}

func TestPrettyPrintedByteCount_Gigabytes(t *testing.T) {
	require.Equal(t, "1.0GB", PrettyPrintedByteCount(1_000_000_000))
	require.Equal(t, "1.2GB", PrettyPrintedByteCount(1_234_567_890))
	require.Equal(t, "500.0GB", PrettyPrintedByteCount(500_000_000_000))
	require.Equal(t, "1000.0GB", PrettyPrintedByteCount(999_999_999_999))
}

func TestCustomStrings_IdentifierToDisaplyName(t *testing.T) {
	require.Equal(t, "Test > One Two Three", IdentifierToDisplayName("test::one-two-three"))
}
