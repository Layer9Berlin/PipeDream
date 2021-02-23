package pipeline

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewDataConnection(t *testing.T) {
	sourceIdentifier := "source"
	targetIdentifier := "target"
	dataConnection := NewDataConnection(&Run{Identifier: &sourceIdentifier}, &Run{Identifier: &targetIdentifier}, "test")
	require.Equal(t, "source", *dataConnection.Source.Identifier)
	require.Equal(t, "target", *dataConnection.Target.Identifier)
	require.Equal(t, "test", *dataConnection.Label)
}
