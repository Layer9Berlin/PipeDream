// Package evaluate provides custom functions extending the `govaluate` package
package evaluate

import (
	"fmt"
	"gopkg.in/Knetic/govaluate.v2"
)

// Bool evaluates an expression that is assumed to be boolean
func Bool(condition string) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return false, fmt.Errorf("error parsing condition %q: %w", condition, err)
	}

	// use empty map instead of nil to prevent panic for certain inputs
	evaluationResult, err := expression.Evaluate(make(map[string]interface{}, 0))
	if err != nil {
		return false, fmt.Errorf("error evaluating condition %q: %w", condition, err)
	}

	evaluatedBoolean, ok := evaluationResult.(bool)
	if !ok {
		return false, fmt.Errorf("condition %q does not evaluate to boolean", condition)
	}

	return evaluatedBoolean, nil
}
