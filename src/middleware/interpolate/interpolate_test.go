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
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
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
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
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
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
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
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
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
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
		"arg":  "value",
		"arg2": "@{arg} @!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
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
		"arg4": "@!!",
		"arg5": "@{missing|default}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("TestInput @{arg}"))

	run.Log.SetLevel(logrus.DebugLevel)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {},
		nil,
	)
	run.Start()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg":  "value",
		"arg2": "test-@{arg}",
		"arg3": "@arg",
		"arg4": "@!!",
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
		"arg": "@!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 2, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "test error")
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "''",
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
	run.Start()
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
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
		"arg": "@{arg2}",
		"arg2": map[string]interface{}{
			"test": "not a valid substitution",
		},
		"arg3": "@!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 2, run.Log.WarnCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
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
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, run.Log.WarnCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg":  "test",
		"arg2": "test",
		"arg3": "@{missing}",
	}, runArguments)
}

func TestInterpolate_EscapeAllQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "all",
			"quote":        "none",
		},
		"arg": "test @!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test \\' \\\" \\` \\\\\" \\'",
	}, runArguments)
}

func TestInterpolate_EscapeSingleQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "single",
			"quote":        "none",
		},
		"arg": "test @!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test \\' \" ` \\\" \\'",
	}, runArguments)
}

func TestInterpolate_EscapeDoubleQuotes(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"escapeQuotes": "double",
			"quote":        "none",
		},
		"arg": "test @!!",
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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test ' \\\" ` \\\\\" '",
	}, runArguments)
}

func TestInterpolate_Nesting(t *testing.T) {
	run, _ := pipeline.NewRun(nil, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"quote": "none",
		},
		"arg1": "value",
		"arg2": "@{arg1}",
		//this should evaluate to "value"
		"arg3": "@{arg4}@{arg5}@{arg6}",
		"arg4": "@{",
		"arg5": "arg2",
		"arg6": "}",
	}, nil, nil)

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
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "value", runArguments["arg3"])
}

func TestInterpolate_QuotingCombinations(t *testing.T) {
	testCases := []map[string]string{
		{
			"escapeQuotes": "single",
		},
		{
			"escapeQuotes": "single",
			"quote":        "none",
		},
		{
			"escapeQuotes": "double",
		},
		{
			"escapeQuotes": "double",
			"quote":        "none",
		},
		{
			"escapeQuotes": "backtick",
		},
		{
			"escapeQuotes": "backtick",
			"quote":        "none",
		},
		{
			"escapeQuotes": "double",
			"quote":        "backtick",
		},
		{
			"escapeQuotes": "single",
			"quote":        "double",
		},
	}
	results := []string{
		"'\"value\" \\\\'with\\\\' `quotes`'",
		"\"value\" \\'with\\' `quotes`",
		"'\\\"value\\\" \\'with\\' `quotes`'",
		"\\\"value\\\" 'with' `quotes`",
		"'\"value\" \\'with\\' \\`quotes\\`'",
		"\"value\" 'with' \\`quotes\\`",
		"`\\\"value\\\" 'with' \\`quotes\\``",
		"\"\\\"value\\\" \\'with\\' `quotes`\"",
	}
	for index, testCase := range testCases {
		run, _ := pipeline.NewRun(nil, map[string]interface{}{
			"interpolate": testCase,
			"arg1":        "\"value\" 'with' `quotes`",
			"arg2":        "@{arg1}",
		}, nil, nil)

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
		run.Start()
		run.Wait()
		waitGroup.Wait()

		require.Equal(t, 0, run.Log.ErrorCount())
		require.Equal(t, results[index], runArguments["arg2"], testCase)
	}
}

func TestInterpolate_PreventInfiniteRecursion_withPipesInterpolation(t *testing.T) {
	runIdentifier := "test"
	preconditionRunIdentifier := "test-precondition"
	run, _ := pipeline.NewRun(&runIdentifier, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"pipes": []string{
				preconditionRunIdentifier,
			},
		},
		"arg": "test @!! @|0",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	executionContext := middleware.NewExecutionContext(
		middleware.WithDefinitionsLookup(pipeline.DefinitionsLookup{
			runIdentifier: []pipeline.Definition{
				{
					DefinitionArguments: map[string]interface{}{
						"interpolate": map[string]interface{}{
							"pipes": []string{
								preconditionRunIdentifier,
							},
						},
					},
				},
			},
		}),
		middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
			defer waitGroup.Done()
			if *childRun.Identifier == "test" {
				runArguments = childRun.ArgumentsCopy()
			}
		}),
	)
	executionContext.FullRun(
		middleware.WithIdentifier(&preconditionRunIdentifier),
		middleware.WithTearDownFunc(func(preconditionRun *pipeline.Run) {
			preconditionRun.Stdout.Replace(strings.NewReader("precondition output"))
		}),
	)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {

		},
		executionContext,
	)
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test 'input' 'precondition output'",
	}, runArguments)
}

func TestInterpolate_PreventInfiniteRecursion_withoutPipesInterpolation(t *testing.T) {
	runIdentifier := "test"
	run, _ := pipeline.NewRun(&runIdentifier, map[string]interface{}{
		"arg": "test @!!",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	executionContext := middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
			defer waitGroup.Done()
			if *childRun.Identifier == "test" {
				runArguments = childRun.ArgumentsCopy()
			}
		}),
	)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {

		},
		executionContext,
	)
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test 'input'",
	}, runArguments)
}

func TestInterpolate_InvalidPipesInterpolation(t *testing.T) {
	runIdentifier := "test"
	preconditionRunIdentifier := "test-precondition"
	run, _ := pipeline.NewRun(&runIdentifier, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"pipes": []string{
				preconditionRunIdentifier,
			},
		},
		"arg": "test @!! @|0 @|2 @|{1} @|{0}",
	}, nil, nil)
	run.Stdin.Replace(strings.NewReader("input"))

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)
	runArguments := make(map[string]interface{}, 0)
	run.Log.SetLevel(logrus.DebugLevel)
	executionContext := middleware.NewExecutionContext(
		middleware.WithExecutionFunction(func(childRun *pipeline.Run) {
			defer waitGroup.Done()
			if *childRun.Identifier == "test" {
				runArguments = childRun.ArgumentsCopy()
			}
		}),
	)
	executionContext.FullRun(
		middleware.WithIdentifier(&preconditionRunIdentifier),
		middleware.WithTearDownFunc(func(preconditionRun *pipeline.Run) {
			preconditionRun.Stdout.Replace(strings.NewReader("precondition output"))
		}),
	)
	NewMiddleware().Apply(
		run,
		func(run *pipeline.Run) {

		},
		executionContext,
	)
	run.Start()
	run.Wait()
	waitGroup.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	// two errors are converted to warnings
	require.Equal(t, 2, run.Log.WarnCount())
	logString := run.Log.String()
	require.Contains(t, logString, "trying to interpolate result at index 2, but only 1 `pipes` argument(s) provided")
	require.Contains(t, logString, "trying to interpolate result at index 1, but only 1 `pipes` argument(s) provided")
	require.Equal(t, map[string]interface{}{
		"interpolate": map[string]interface{}{
			"enable": false,
		},
		"arg": "test 'input' 'precondition output' '' '' 'precondition output'",
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
