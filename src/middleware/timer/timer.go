// Package timer provides a middleware that records execution time
package timer

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"time"
)

// Execution Time Recorder
type timeProvider interface {
	Now() time.Time
	Since(time time.Time) time.Duration
}

type defaultTimeProvider struct{}

func (defaultTimeProvider) Now() time.Time {
	return time.Now()
}

func (defaultTimeProvider) Since(startTime time.Time) time.Duration {
	return time.Since(startTime)
}

// Middleware is an execution time recorder
type Middleware struct {
	timeProvider timeProvider
}

type timerMiddlewareArguments struct {
	Record bool
}

// String is a human-readable description
func (timerMiddleware Middleware) String() string {
	return "timer"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return NewMiddlewareWithProvider(defaultTimeProvider{})
}

// NewMiddlewareWithProvider creates a new middleware instance with the specified time provider
func NewMiddlewareWithProvider(timeProvider timeProvider) Middleware {
	return Middleware{
		timeProvider: timeProvider,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (timerMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := timerMiddlewareArguments{
		Record: false,
	}
	pipeline.ParseArguments(&arguments, "timer", run)

	if arguments.Record {
		run.Log.Trace(
			fields.Symbol("🕑"),
			fields.Message("starting execution timer..."),
			fields.Middleware(timerMiddleware),
		)
		start := timerMiddleware.timeProvider.Now()
		next(run)
		recordDuration := timerMiddleware.timeProvider.Since(start)
		run.Log.Debug(
			fields.Symbol("🕑"),
			fields.Message("execution time"),
			fields.Info(recordDuration),
			fields.Middleware(timerMiddleware),
		)
	} else {
		next(run)
	}
}
