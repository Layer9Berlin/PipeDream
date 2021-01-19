package timer

import (
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
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
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	arguments := timerMiddlewareArguments{
		Record: false,
	}
	middleware.ParseArguments(&arguments, "timer", run)

	if arguments.Record {
		run.Log.TraceWithFields(
			log_fields.Symbol("🕑"),
			log_fields.Message("starting execution timer..."),
			log_fields.Middleware(timerMiddleware),
		)
		start := timerMiddleware.timeProvider.Now()
		next(run)
		recordDuration := timerMiddleware.timeProvider.Since(start)
		run.Log.DebugWithFields(
			log_fields.Symbol("🕑"),
			log_fields.Message("execution time"),
			log_fields.Info(recordDuration),
			log_fields.Middleware(timerMiddleware),
		)
	} else {
		next(run)
	}
}
