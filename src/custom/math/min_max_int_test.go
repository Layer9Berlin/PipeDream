package math

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCustomMath_MaxInt(t *testing.T) {
	require.Equal(t, 1, MaxInt(0, 1))
	require.Equal(t, 3, MaxInt(3, 3))
	require.Equal(t, 4, MaxInt(4, -5))
}

func TestCustomMath_MinInt(t *testing.T) {
	require.Equal(t, 0, MinInt(0, 1))
	require.Equal(t, 3, MinInt(3, 3))
	require.Equal(t, -5, MinInt(4, -5))
}
