package version

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion_Cmd(t *testing.T) {
	var buffer bytes.Buffer

	version = "1.0.0"
	RepoChecksum = "test-checksum"

	Cmd(&buffer)

	require.Contains(t, buffer.String(), "1.0.0 (repo checksum: test-checksum)")
}
