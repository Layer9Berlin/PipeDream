package middleware

import (
	"github.com/Layer9Berlin/pipedream/src/models"
)

type Middleware interface {
	String() string
	Apply(
		run *models.PipelineRun,
		next func(*models.PipelineRun),
		executionContext *ExecutionContext,
	)
}
