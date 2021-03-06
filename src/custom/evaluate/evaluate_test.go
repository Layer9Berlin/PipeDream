package evaluate

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNonParseableCondition(t *testing.T) {
	_, err := Bool("(")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "error parsing condition")
}

func TestNonEvaluableCondition(t *testing.T) {
	_, err := Bool("test")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "error evaluating condition")
}

func TestNonBooleanCondition(t *testing.T) {
	_, err := Bool("\"test\"")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "does not evaluate to boolean")
}

func TestBooleanCondition(t *testing.T) {
	result, err := Bool("true")
	require.Nil(t, err)
	require.Equal(t, result, true)
}
