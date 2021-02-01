package interpolate

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/ryankurte/go-structparse"
	"io/ioutil"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Value Replacer
type interpolateMiddlewareArguments struct {
	Enable         bool
	EscapeQuotes   string
	IgnoreWarnings bool
	Quote          string
}

type InterpolateMiddleware struct{}

func (interpolateMiddleware InterpolateMiddleware) String() string {
	return "interpolate"
}

func NewInterpolateMiddleware() InterpolateMiddleware {
	return InterpolateMiddleware{}
}

func (interpolateMiddleware InterpolateMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := interpolateMiddlewareArguments{
		Enable:         true,
		EscapeQuotes:   "none",
		IgnoreWarnings: false,
		Quote:          "none",
	}
	middleware.ParseArguments(&arguments, "interpolate", run)

	next(run)

	if arguments.Enable {
		interpolator := NewInterpolator(run.ArgumentsCopy(), arguments)

		interpolatedArguments := run.ArgumentsCopy()
		structparse.Strings(interpolator, interpolatedArguments)

		if interpolator.NeedCompleteInput {
			// we log any errors only as warnings
			// this is because we might have other middleware (like `when`)
			// that renders certain errors moot
			if interpolator.Errors != nil && interpolator.Errors.Len() > 0 {
				if !arguments.IgnoreWarnings {
					run.Log.WarnWithFields(
						log_fields.Symbol("⚠️"),
						log_fields.Message("warning"),
						log_fields.Info(interpolator.Errors.Errors),
						log_fields.Middleware(interpolateMiddleware),
					)
				}
			} else {
				run.Log.DebugWithFields(
					log_fields.Symbol("💤"),
					log_fields.Message("input interpolation used, need to wait for input to complete..."),
					log_fields.Middleware(interpolateMiddleware),
				)
			}
			run.Log.TraceWithFields(
				log_fields.DataStream(interpolateMiddleware, "copying stdin")...,
			)
			stdinCopy := run.Stdin.CopyOrResult()
			run.Log.TraceWithFields(
				log_fields.DataStream(interpolateMiddleware, "creating stdout writer")...,
			)
			stdoutAppender := run.Stdout.WriteCloser()
			run.Log.TraceWithFields(
				log_fields.DataStream(interpolateMiddleware, "creating stderr writer")...,
			)
			stderrAppender := run.Stdout.WriteCloser()
			// we return immediately and wait for the previous input to be available
			// then we execute a full run
			parentLogWriter := run.Log.AddWriteCloserEntry()
			go func() {
				input, inputErr := ioutil.ReadAll(stdinCopy)
				interpolator := NewInterpolatorWithInput(interpolatedArguments, arguments, input)
				structparse.Strings(interpolator, interpolatedArguments)
				executionContext.FullRun(
					middleware.WithIdentifier(run.Identifier),
					middleware.WithParentRun(run),
					middleware.WithLogWriter(parentLogWriter),
					middleware.WithArguments(interpolatedArguments),
					middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
						interpolator.log(childRun.Log, interpolateMiddleware)
						childRun.Log.PossibleErrorWithExplanation(inputErr, "unable to find value for previous output")
						childRun.Log.TraceWithFields(
							log_fields.DataStream(interpolateMiddleware, "merging parent stdin into child stdin")...,
						)
						childRun.Stdin.MergeWith(bytes.NewReader(input))
					}),
					middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
						childRun.Log.TraceWithFields(
							log_fields.DataStream(interpolateMiddleware, "merging child stdout into parent stdout")...,
						)
						childRun.Stdout.StartCopyingInto(stdoutAppender)
						childRun.Log.TraceWithFields(
							log_fields.DataStream(interpolateMiddleware, "merging child stderr into parent stderr")...,
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

		interpolator.log(run.Log, interpolateMiddleware)
		if len(interpolator.Substitutions) > 0 {
			run.Log.TraceWithFields(
				log_fields.DataStream(interpolateMiddleware, "creating parent stdout writer")...,
			)
			stdoutAppender := run.Stdout.WriteCloser()
			run.Log.TraceWithFields(
				log_fields.DataStream(interpolateMiddleware, "creating parent stderr writer")...,
			)
			stderrAppender := run.Stderr.WriteCloser()
			executionContext.FullRun(
				middleware.WithIdentifier(run.Identifier),
				middleware.WithParentRun(run),
				middleware.WithArguments(interpolatedArguments),
				middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
					childRun.Log.TraceWithFields(
						log_fields.DataStream(interpolateMiddleware, "merging parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(run.Stdin.CopyOrResult())
				}),
				middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
					childRun.Log.TraceWithFields(
						log_fields.DataStream(interpolateMiddleware, "merging child stdout into parent stdout writer")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutAppender)
					childRun.Log.TraceWithFields(
						log_fields.DataStream(interpolateMiddleware, "merging child stderr into parent stderr writer")...,
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
}

type Interpolator struct {
	ArgumentReplacements map[string]interface{}
	completeInput        []byte
	Errors               *multierror.Error
	MiddlewareArguments  interpolateMiddlewareArguments
	NeedCompleteInput    bool
	Substitutions        map[string]interface{}
	WaitGroup            *sync.WaitGroup
}

func NewInterpolatorWithInput(
	availableReplacements map[string]interface{},
	middlewareArguments interpolateMiddlewareArguments,
	completeInput []byte,
) *Interpolator {
	interpolator := Interpolator{
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

func NewInterpolator(
	availableReplacements map[string]interface{},
	middlewareArguments interpolateMiddlewareArguments,
) *Interpolator {
	return NewInterpolatorWithInput(availableReplacements, middlewareArguments, nil)
}

func (interpolator *Interpolator) ParseString(value string) interface{} {
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

func (interpolator *Interpolator) interpolateInput(value string) string {
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

func (interpolator *Interpolator) interpolateArguments(value string) (string, error) {
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
		} else {
			replacement = "false"
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
			} else {
				// note that match[3] might be "", but we do want to allow expressions
				// of the form `@{key|}` that do not throw an error if the value can't be found
				interpolator.Substitutions[key] = match[3]
				value = strings.Replace(value, match[0], match[3], 1)
			}
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

func (interpolator *Interpolator) log(logger *models.PipelineRunLogger, interpolateMiddleware InterpolateMiddleware) {
	if interpolator.Errors != nil && interpolator.Errors.Len() > 0 {
		if !interpolator.MiddlewareArguments.IgnoreWarnings {
			logger.WarnWithFields(
				log_fields.Symbol("⚠️"),
				log_fields.Message("warning"),
				log_fields.Info(interpolator.Errors.Errors),
				log_fields.Middleware(interpolateMiddleware),
			)
		}
	} else {
		switch len(interpolator.Substitutions) {
		case 0:
		case 1:
			logger.DebugWithFields(
				log_fields.Symbol("⎆"),
				log_fields.Message("made 1 substitution"),
				log_fields.Info(interpolator.Substitutions),
				log_fields.Middleware(interpolateMiddleware),
			)
		default:
			logger.DebugWithFields(
				log_fields.Symbol("⎆"),
				log_fields.Message(fmt.Sprintf("made %v substitutions", len(interpolator.Substitutions))),
				log_fields.Info(interpolator.Substitutions),
				log_fields.Middleware(interpolateMiddleware),
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
