package version

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion_Cmd(t *testing.T) {
	var buffer bytes.Buffer

	Version = "1.0.0"
	RepoChecksum = "test-checksum"

	Cmd(&buffer)

	require.Contains(t, buffer.String(), "version: 1.0.0")
	require.Contains(t, buffer.String(), "checksum: test-checksum")
}
