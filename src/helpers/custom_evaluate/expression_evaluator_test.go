package custom_evaluate

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNonParseableCondition(t *testing.T) {
	_, err := EvaluateBool("(")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "error parsing condition")
}

func TestNonEvaluableCondition(t *testing.T) {
	_, err := EvaluateBool("test")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "error evaluating condition")
}

func TestNonBooleanCondition(t *testing.T) {
	_, err := EvaluateBool("\"test\"")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "does not custom_evaluate to boolean")
}

func TestBooleanCondition(t *testing.T) {
	result, err := EvaluateBool("true")
	require.Nil(t, err)
	require.Equal(t, result, true)
}
