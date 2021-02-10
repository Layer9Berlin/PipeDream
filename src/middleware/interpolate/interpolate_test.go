package interpolate

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestInterpolate_ArgumentSubstitution(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "test @{arg}",
		},
		"arg":   "value",
		"arg2":  "test-@{arg}",
		"arg3":  "@arg",
		"arg5":  "@{missing|default}",
		"arg6":  []interface{}{"line0", "line1", "line2"},
		"arg7":  13,
		"arg8":  "@?{arg}",
		"arg9":  "@?{missing}",
		"arg10": "@{arg6}",
		"arg11": "@{arg7}",
		"arg12": "@{missing|}",
	}, nil, nil)
	// arguments in the input will not themselves be interpolated
	// but the input might be interpolated into arguments
	// and substitutions will be made there
	run.Stdin.Replace(strings.NewReader("TestInput @{arg}"))

	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 0)
	childIdentifier := ""
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				childRun.Log.Debug(fields.Message("child run log entry"))
				childIdentifier = *childRun.Identifier
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"shell": map[string]interface{}{
			"run": "test value",
		},
		"arg":   "value",
		"arg2":  "test-value",
		"arg3":  "@arg",
		"arg5":  "default",
		"arg6":  []interface{}{"line0", "line1", "line2"},
		"arg7":  13,
		"arg8":  "true",
		"arg9":  "false",
		"arg10": "line0\nline1\nline2",
		"arg11": "13",
		"arg12": "",
	}, runArguments)
	require.Equal(t, "child identifier with @{arg} (not interpolated)", childIdentifier)
	logString := run.Log.String()
	require.Contains(t, logString, "child run log entry")
	require.Equal(t, "TestInput @{arg}", run.Stdin.String())
	require.Contains(t, logString, "made 6 substitutions")
}

func TestInterpolate_SingleSubstitution(t *testing.T) {
	identifier := "child identifier"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg":  "value",
		"arg2": "@{arg}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("TestInput"))

	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 0)
	childIdentifier := ""
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				childRun.Log.Debug(fields.Message("child run log entry"))
				childIdentifier = *childRun.Identifier
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"arg":  "value",
		"arg2": "value",
	}, runArguments)
	require.Equal(t, "child identifier", childIdentifier)
	logString := run.Log.String()
	require.Contains(t, logString, "child run log entry")
	require.Contains(t, logString, "made 1 substitution")
}

func TestInterpolate_InputAndArgumentSubstitution(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg":  "value",
		"arg2": "@{arg} $!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("TestInput @{arg}"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 2)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"arg":  "value",
		"arg2": "value TestInput value",
	}, runArguments)
	require.Equal(t, "child identifier with @{arg} (not interpolated)", *run.Identifier)
	require.Contains(t, run.Log.String(), "made 2 substitutions")
}

func TestInterpolate_Disabled(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg":  "value",
		"arg2": "test-@{arg}",
		"arg3": "@arg",
		"arg4": "$!!",
		"arg5": "@{missing|default}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("TestInput @{arg}"))

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg":  "value",
		"arg2": "test-@{arg}",
		"arg3": "@arg",
		"arg4": "$!!",
		"arg5": "@{missing|default}",
	}, run.ArgumentsCopy())
	require.NotContains(t, run.Log.String(), "interpolate")
}

func TestInterpolate_NoSubstitution(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg": "value",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("TestInput @{arg}"))

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"arg": "value",
	}, run.ArgumentsCopy())
}

func TestInterpolate_InputReadError(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg": "$!!",
	}, nil, nil)
	run.Stdin.Replace(NewErrorReader(1))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 2)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "test error")
	require.Equal(t, map[string]interface{}{
		"arg": "",
	}, runArguments)
}

func TestInterpolate_ValueMissing(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg": "@{missing}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, run.Log.WarnCount())
	require.Equal(t, map[string]interface{}{
		"arg": "@{missing}",
	}, run.ArgumentsCopy())
}

func TestInterpolate_ValueNotSubstitutable(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg": "@{arg2}",
		"arg2": map[string]interface{}{
			"test": "not a valid substitution",
		},
		"arg3": "$!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 2)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, run.Log.WarnCount())
	require.Equal(t, map[string]interface{}{
		"arg": "@{arg2}",
		"arg2": map[string]interface{}{
			"test": "not a valid substitution",
		},
		"arg3": "input",
	}, runArguments)
}

func TestInterpolate_SubstitutionPlusError(t *testing.T) {
	identifier := "child identifier with @{arg} (not interpolated)"
	run, _ := pipeline.NewRun(&identifier, map[string]interface{}{
		"arg":  "@{arg2}",
		"arg2": "test",
		"arg3": "@{missing}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	run.Log.SetLevel(logrus.DebugLevel)
	runArguments := make(map[string]interface{}, 2)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, run.Log.WarnCount())
	require.Equal(t, map[string]interface{}{
		"arg":  "test",
		"arg2": "test",
		"arg3": "@{missing}",
	}, runArguments)
}

func TestInterpolate_EscapeAllQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "all",
		},
		"arg": "test $!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("' \" ` \\\" '"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "all",
		},
		"arg": "test \\\" \\\" ` \\\\\" \\\"",
	}, runArguments)
}

func TestInterpolate_EscapeSingleQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "single",
		},
		"arg": "test $!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("' \" ` \\\" '"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "single",
		},
		"arg": "test \\\" \" ` \\\" \\\"",
	}, runArguments)
}

func TestInterpolate_EscapeDoubleQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "double",
		},
		"arg": "test $!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("' \" ` \\\" '"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		middleware.NewExecutionContext(
			middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
				defer waitGroup.Done()
				runArguments = childRun.ArgumentsCopy()
			}),
		))
	run.Close()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "double",
		},
		"arg": "test ' \\\" ` \\\\\" '",
	}, runArguments)
}

type ErrorReader struct {
	counter int
}

func NewErrorReader(counter int) *ErrorReader {
	return &ErrorReader{
		counter: counter,
	}
}

func (errorWriter *ErrorReader) Read(_ []byte) (int, error) {
	if errorWriter.counter <= 0 {
		return 0, io.EOF
	}
	errorWriter.counter = errorWriter.counter - 1
	return 0, fmt.Errorf("test error")
}
