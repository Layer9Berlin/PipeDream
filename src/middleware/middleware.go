// Package middleware slices the implementation into middlewares, each performing side effects and adapting the run based on provided arguments
package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
)

type Middleware interface {
	String() string
	Apply(
		run *pipeline.Run,
		next func(*pipeline.Run),
		executionContext *ExecutionContext,
	)
}
