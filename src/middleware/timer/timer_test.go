package timer

import (
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type MockTimeProvider struct {
	Times []time.Time
}

func (timeProvider *MockTimeProvider) Now() time.Time {
	currentTime, remainingTimes := timeProvider.Times[0], timeProvider.Times[1:]
	timeProvider.Times = remainingTimes
	return currentTime
}

func (timeProvider *MockTimeProvider) Since(startTime time.Time) time.Duration {
	currentTime, remainingTimes := timeProvider.Times[0], timeProvider.Times[1:]
	timeProvider.Times = remainingTimes
	return currentTime.Sub(startTime)
}

func TestTimer_WithValidArguments_RecordsExecutionTime(t *testing.T) {
	timeProvider := &MockTimeProvider{
		Times: []time.Time{
			time.Unix(1606652605, 0),
			time.Unix(1606652605, 10_000_000),
		},
	}

	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"timer": map[string]interface{}{
			"record": true,
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddlewareWithProvider(timeProvider).Apply(run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "timer | execution time | 10ms")
}

func TestTimer_WithInvalidArguments_ThrowsError(t *testing.T) {
	timeProvider := MockTimeProvider{
		Times: []time.Time{
			time.Unix(1606652605, 0),
			time.Unix(1606652605, 10_000_000),
		},
	}

	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"timer": []string{
			"invalid",
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddlewareWithProvider(&timeProvider).Apply(run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
	require.NotContains(t, run.Log.String(), "execution time")
}

func TestTimer_NoArguments_DeactivateTimer(t *testing.T) {
	timeProvider := MockTimeProvider{
		Times: []time.Time{
			time.Unix(1606652605, 0),
			time.Unix(1606652605, 10_000_000),
		},
	}

	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, nil, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddlewareWithProvider(&timeProvider).Apply(run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.NotContains(t, run.Log.String(), "execution time")
}

func TestDefaultTimeProvider(t *testing.T) {
	identifier := "test"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"timer": map[string]interface{}{
			"record": true,
		},
	}, nil, nil)

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Contains(t, run.Log.String(), "execution time")
}
