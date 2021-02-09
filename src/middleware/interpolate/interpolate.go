// Package interpolate provides a middleware to substitute arguments or inputs into other arguments
package interpolate

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/hashicorp/go-multierror"
	"github.com/ryankurte/go-structparse"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type interpolateMiddlewareArguments struct {
	Enable         bool
	EscapeQuotes   string
	IgnoreWarnings bool
	Quote          string
}

// Middleware is a value replacer
type Middleware struct{}

// String is a human-readable description
func (interpolateMiddleware Middleware) String() string {
	return "interpolate"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (interpolateMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	arguments := interpolateMiddlewareArguments{
		Enable:         true,
		EscapeQuotes:   "none",
		IgnoreWarnings: false,
		Quote:          "none",
	}
	pipeline.ParseArguments(&arguments, "interpolate", run)

	if arguments.Enable {
		preliminaryInterpolator := newInterpolator(run.ArgumentsCopy(), arguments)

		interpolatedArguments := run.ArgumentsCopy()
		structparse.Strings(preliminaryInterpolator, interpolatedArguments)

		if preliminaryInterpolator.NeedCompleteInput {
			// we log any errors only as warnings
			// this is because we might have other middleware (like `when`)
			// that renders certain errors moot
			if preliminaryInterpolator.Errors != nil && preliminaryInterpolator.Errors.Len() > 0 {
				if !arguments.IgnoreWarnings {
					run.Log.Warn(
						fields.Symbol("⚠️"),
						fields.Message("warning"),
						fields.Info(preliminaryInterpolator.Errors.Errors),
						fields.Middleware(interpolateMiddleware),
					)
				}
			} else {
				run.Log.Debug(
					fields.Symbol("💤"),
					fields.Message("input interpolation used, need to wait for input to complete..."),
					fields.Middleware(interpolateMiddleware),
				)
			}
			run.Log.Trace(
				fields.DataStream(interpolateMiddleware, "copying stdin")...,
			)
			stdinCopy := run.Stdin.Copy()
			run.Log.Trace(
				fields.DataStream(interpolateMiddleware, "creating stdout writer")...,
			)
			stdoutAppender := run.Stdout.WriteCloser()
			run.Log.Trace(
				fields.DataStream(interpolateMiddleware, "creating stderr writer")...,
			)
			stderrAppender := run.Stdout.WriteCloser()
			// we return immediately and wait for the previous input to be available
			// then we execute a full run
			parentLogWriter := run.Log.AddWriteCloserEntry()
			go func() {
				input, inputErr := ioutil.ReadAll(stdinCopy)
				fullInterpolator := newInterpolatorWithInput(interpolatedArguments, arguments, input)
				structparse.Strings(fullInterpolator, interpolatedArguments)
				executionContext.FullRun(
					middleware.WithIdentifier(run.Identifier),
					middleware.WithParentRun(run),
					middleware.WithLogWriter(parentLogWriter),
					middleware.WithArguments(interpolatedArguments),
					middleware.WithSetupFunc(func(childRun *pipeline.Run) {
						fullInterpolator.log(childRun.Log, interpolateMiddleware)
						childRun.Log.PossibleErrorWithExplanation(inputErr, "unable to find value for previous output")
						childRun.Log.Trace(
							fields.DataStream(interpolateMiddleware, "merging parent stdin into child stdin")...,
						)
						childRun.Stdin.MergeWith(bytes.NewReader(input))
					}),
					middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
						childRun.Log.Trace(
							fields.DataStream(interpolateMiddleware, "merging child stdout into parent stdout")...,
						)
						childRun.Stdout.StartCopyingInto(stdoutAppender)
						childRun.Log.Trace(
							fields.DataStream(interpolateMiddleware, "merging child stderr into parent stderr")...,
						)
						childRun.Stderr.StartCopyingInto(stderrAppender)
						go func() {
							childRun.Wait()
							// need to clean up by closing the writers we created
							childRun.Log.PossibleError(stdoutAppender.Close())
							childRun.Log.PossibleError(stderrAppender.Close())
						}()
					}))
			}()
			return
		}

		preliminaryInterpolator.log(run.Log, interpolateMiddleware)
		if len(preliminaryInterpolator.Substitutions) > 0 {
			run.Log.Trace(
				fields.DataStream(interpolateMiddleware, "creating parent stdout writer")...,
			)
			stdoutAppender := run.Stdout.WriteCloser()
			run.Log.Trace(
				fields.DataStream(interpolateMiddleware, "creating parent stderr writer")...,
			)
			stderrAppender := run.Stderr.WriteCloser()
			executionContext.FullRun(
				middleware.WithIdentifier(run.Identifier),
				middleware.WithParentRun(run),
				middleware.WithArguments(interpolatedArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					childRun.Log.Trace(
						fields.DataStream(interpolateMiddleware, "merging parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(run.Stdin.Copy())
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					childRun.Log.Trace(
						fields.DataStream(interpolateMiddleware, "merging child stdout into parent stdout writer")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutAppender)
					childRun.Log.Trace(
						fields.DataStream(interpolateMiddleware, "merging child stderr into parent stderr writer")...,
					)
					childRun.Stderr.StartCopyingInto(stderrAppender)
					go func() {
						childRun.Wait()
						// need to clean up by closing the writers we created
						run.Log.PossibleError(stdoutAppender.Close())
						run.Log.PossibleError(stderrAppender.Close())
					}()
				}))
			return
		}
	}

	next(run)
}

type interpolator struct {
	ArgumentReplacements map[string]interface{}
	completeInput        []byte
	Errors               *multierror.Error
	MiddlewareArguments  interpolateMiddlewareArguments
	NeedCompleteInput    bool
	Substitutions        map[string]interface{}
	WaitGroup            *sync.WaitGroup
}

func newInterpolatorWithInput(
	availableReplacements map[string]interface{},
	middlewareArguments interpolateMiddlewareArguments,
	completeInput []byte,
) *interpolator {
	interpolator := interpolator{
		ArgumentReplacements: availableReplacements,
		completeInput:        completeInput,
		Errors:               nil,
		MiddlewareArguments:  middlewareArguments,
		NeedCompleteInput:    false,
		Substitutions:        make(map[string]interface{}, 10),
		WaitGroup:            &sync.WaitGroup{},
	}
	return &interpolator
}

func newInterpolator(
	availableReplacements map[string]interface{},
	middlewareArguments interpolateMiddlewareArguments,
) *interpolator {
	return newInterpolatorWithInput(availableReplacements, middlewareArguments, nil)
}

func (interpolator *interpolator) ParseString(value string) interface{} {
	// we allow defining keys in terms of arguments and vice versa, within reason
	result := interpolator.interpolateInput(value)

	substitutionCount := len(interpolator.Substitutions)
	// keep replacing arguments while all three conditions are fulfilled:
	// - the total number of iterations is at most 5
	// - no errors have been encountered
	// - the last iteration made a substitution
	for range [5]int{} {
		var err error
		result, err = interpolator.interpolateArguments(result)
		if err != nil {
			interpolator.Errors = multierror.Append(interpolator.Errors, err)
			break
		}
		if len(interpolator.Substitutions) == substitutionCount {
			break
		}
		substitutionCount = len(interpolator.Substitutions)
	}

	return result
}

func (interpolator *interpolator) interpolateInput(value string) string {
	if strings.Contains(value, "$!!") {
		interpolator.NeedCompleteInput = true
		if interpolator.completeInput != nil {
			replacement := escapeQuotes(string(interpolator.completeInput), interpolator.MiddlewareArguments)
			interpolator.Substitutions["$!!"] = replacement
			return strings.Replace(value, "$!!", replacement, -1)
		}
	}
	return value
}

func (interpolator *interpolator) interpolateArguments(value string) (string, error) {
	// we proceed even if we have no valid replacements
	// as @? directives should still be processed
	// and default values substituted
	regex := regexp.MustCompile("@\\?{([0-9a-zA-Z\\-_.:/\\\\]*)?}")
	matches := regex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		key := match[1]
		argument, haveArgument := interpolator.ArgumentReplacements[key]
		replacement := "false"
		// allow unsetting of parent options using `null`
		if haveArgument && argument != nil {
			replacement = "true"
		}
		interpolator.Substitutions[fmt.Sprintf("have %v", key)] = replacement
		value = strings.Replace(value, match[0], replacement, 1)
	}

	regex = regexp.MustCompile("@{([0-9a-zA-Z\\-_.:/\\\\]*)( *\\| *([0-9a-zA-Z-_.:'\"]*))?}")
	matches = regex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		key := match[1]
		replacement, haveReplacement := interpolator.ArgumentReplacements[key]
		if !haveReplacement || replacement == nil {
			if match[2] == "" {
				// not finding a value does not necessarily indicate an error
				// we might not end up executing the respective part of the pipe
				return value, fmt.Errorf("unable to find value for argument: `%v` (this might be fine)", key)
			}

			// note that match[3] might be "", but we do want to allow expressions
			// of the form `@{key|}` that do not throw an error if the value can't be found
			interpolator.Substitutions[key] = match[3]
			value = strings.Replace(value, match[0], match[3], 1)
		} else {
			switch typedReplacement := replacement.(type) {
			case string:
				interpolator.Substitutions[key] = typedReplacement
				value = strings.Replace(value, match[0], typedReplacement, 1)
			case []interface{}:
				stringValues := make([]string, 0, len(typedReplacement))
				for _, optionValue := range typedReplacement {
					stringOptionValue, ok := optionValue.(string)
					if ok {
						stringValues = append(stringValues, stringOptionValue)
					}
				}
				stringReplacement := strings.Join(stringValues, "\n")
				interpolator.Substitutions[key] = stringReplacement
				value = strings.Replace(value, match[0], stringReplacement, 1)
			case int:
				stringReplacement := strconv.Itoa(typedReplacement)
				interpolator.Substitutions[key] = stringReplacement
				value = strings.Replace(value, match[0], stringReplacement, 1)
			default:
				return value, fmt.Errorf("value for argument `%v` is not a string, int or array of strings", key)
			}
		}
	}
	return value, nil
}

func (interpolator *interpolator) log(logger *pipeline.Logger, interpolateMiddleware Middleware) {
	if interpolator.Errors != nil && interpolator.Errors.Len() > 0 {
		if !interpolator.MiddlewareArguments.IgnoreWarnings {
			logger.Warn(
				fields.Symbol("⚠️"),
				fields.Message("warning"),
				fields.Info(interpolator.Errors.Errors),
				fields.Middleware(interpolateMiddleware),
			)
		}
	} else {
		switch len(interpolator.Substitutions) {
		case 0:
		case 1:
			logger.Debug(
				fields.Symbol("⎆"),
				fields.Message("made 1 substitution"),
				fields.Info(interpolator.Substitutions),
				fields.Middleware(interpolateMiddleware),
			)
		default:
			logger.Debug(
				fields.Symbol("⎆"),
				fields.Message(fmt.Sprintf("made %v substitutions", len(interpolator.Substitutions))),
				fields.Info(interpolator.Substitutions),
				fields.Middleware(interpolateMiddleware),
			)
		}
	}
}

func escapeQuotes(message string, arguments interpolateMiddlewareArguments) string {
	switch arguments.EscapeQuotes {
	case "all":
		return strings.Replace(strings.Replace(message, "\"", "\\\"", -1), "'", "\\\"", -1)
	case "double":
		return strings.Replace(message, "\"", "\\\"", -1)
	case "single":
		return strings.Replace(message, "'", "\\\"", -1)
	default:
		return message
	}
}
