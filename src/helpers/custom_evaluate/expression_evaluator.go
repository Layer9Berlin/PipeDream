package custom_evaluate

import (
	"fmt"
	"gopkg.in/Knetic/govaluate.v2"
)

func EvaluateBool(condition string) (bool, error) {
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
		return false, fmt.Errorf("condition %q does not custom_evaluate to boolean", condition)
	}

	return evaluatedBoolean, nil
}
