package custom_filepath

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAbsolutePathWithAbsolutePath(t *testing.T) {
	result, err := AbsolutePath("/test/path", "/another/test/path")
	require.Equal(t, "/test/path", result)
	require.Nil(t, err)
}

func TestAbsolutePathWithRelativePath(t *testing.T) {
	result, err := AbsolutePath("test/path", "/another/test/path")
	require.Equal(t, "/another/test/path/test/path", result)
	require.Nil(t, err)
}
