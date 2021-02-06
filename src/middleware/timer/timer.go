// The `timer` middleware records execution time
package timer

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"time"
)

// Execution Time Recorder
type TimeProvider interface {
	Now() time.Time
	Since(time time.Time) time.Duration
}

type defaultTimeProvider struct{}

func (_ defaultTimeProvider) Now() time.Time {
	return time.Now()
}

func (_ defaultTimeProvider) Since(startTime time.Time) time.Duration {
	return time.Since(startTime)
}

type TimerMiddleware struct {
	timeProvider TimeProvider
}

type timerMiddlewareArguments struct {
	Record bool
}

func (timerMiddleware TimerMiddleware) String() string {
	return "timer"
}

func NewTimerMiddleware() TimerMiddleware {
	return NewTimerMiddlewareWithProvider(defaultTimeProvider{})
}

func NewTimerMiddlewareWithProvider(timeProvider TimeProvider) TimerMiddleware {
	return TimerMiddleware{
		timeProvider: timeProvider,
	}
}

func (timerMiddleware TimerMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := timerMiddlewareArguments{
		Record: false,
	}
	middleware.ParseArguments(&arguments, "timer", run)

	if arguments.Record {
		run.Log.TraceWithFields(
			fields.Symbol("🕑"),
			fields.Message("starting execution timer..."),
			fields.Middleware(timerMiddleware),
		)
		start := timerMiddleware.timeProvider.Now()
		next(run)
		recordDuration := timerMiddleware.timeProvider.Since(start)
		run.Log.DebugWithFields(
			fields.Symbol("🕑"),
			fields.Message("execution time"),
			fields.Info(recordDuration),
			fields.Middleware(timerMiddleware),
		)
	} else {
		next(run)
	}
}
