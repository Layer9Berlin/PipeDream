// Package middleware provides middlewares that together implement pipedream's core functionalities
package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

// Middleware slices the implementation into middlewares, each performing side effects and adapting the run based on provided arguments
//
// each interpret a slice of the run's arguments and transform provide
//transforms the run during execution
type Middleware interface {
	String() string
	Apply(
		run *pipeline.Run,
		next func(*pipeline.Run),
		executionContext *ExecutionContext,
	)
}
