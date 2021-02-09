// Package env provides a middleware handling environment variables
package env

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/ryankurte/go-structparse"
	"os"
)

type envMiddlewareArguments struct {
	Interpolate string
	Save        *string
}

// Middleware is an environment variable handler
type Middleware struct {
	Setenv    func(string, string) error
	ExpandEnv func(string) string
}

// String is a human-readable description
func (envMiddleware Middleware) String() string {
	return "env"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return NewMiddlewareWithProvider(os.Setenv, os.ExpandEnv)
}

// NewMiddlewareWithProvider creates a new middleware instance with the specified env functions
func NewMiddlewareWithProvider(
	setenv func(string, string) error,
	expandEnv func(string) string,
) Middleware {
	return Middleware{
		Setenv:    setenv,
		ExpandEnv: expandEnv,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (envMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := envMiddlewareArguments{
		Interpolate: "shallow",
		Save:        nil,
	}
	pipeline.ParseArguments(&arguments, "env", run)

	envInterpolator := newInterpolator(envMiddleware)
	interpolatedArguments := run.ArgumentsCopy()
	switch arguments.Interpolate {
	case "deep":
		structparse.Strings(envInterpolator, interpolatedArguments)
		run.SetArguments(interpolatedArguments)
	case "none":
	default:
		newArguments := run.ArgumentsCopy()
		for argumentKey, argumentValue := range interpolatedArguments {
			if argumentValueAsString, argumentValueIsString := argumentValue.(string); argumentValueIsString {
				newArguments[argumentKey] = envInterpolator.ParseString(argumentValueAsString)
			}
		}
		run.SetArguments(newArguments)
	}
	switch len(envInterpolator.Substitutions) {
	case 0:
	case 1:
		run.Log.Debug(
			fields.Symbol("💲"),
			fields.Message("made 1 env var substitution"),
			fields.Info(envInterpolator.Substitutions),
			fields.Middleware(envMiddleware),
		)
	default:
		run.Log.Debug(
			fields.Symbol("💲"),
			fields.Message(fmt.Sprintf("made %v env var substitutions", len(envInterpolator.Substitutions))),
			fields.Info(envInterpolator.Substitutions),
			fields.Middleware(envMiddleware),
		)
	}

	next(run)

	if arguments.Save != nil {
		run.LogClosingWaitGroup.Add(1)
		go func() {
			run.Stdout.Wait()
			err := envMiddleware.Setenv(*arguments.Save, run.Stdout.String())
			run.Log.PossibleError(err)
			run.LogClosingWaitGroup.Done()
		}()
		run.Log.Debug(
			fields.Symbol("💲"),
			fields.Message(fmt.Sprintf("saving output")),
			fields.Info("$"+*arguments.Save),
			fields.Middleware(envMiddleware),
		)
	}
}

type interpolator struct {
	Substitutions map[string]interface{}
	ExpandEnv     func(string) string
}

func newInterpolator(envMiddleware Middleware) *interpolator {
	return &interpolator{
		Substitutions: make(map[string]interface{}, 10),
		ExpandEnv:     envMiddleware.ExpandEnv,
	}
}

func (interpolator *interpolator) ParseString(value string) interface{} {
	result := interpolator.ExpandEnv(value)
	if result != value {
		interpolator.Substitutions[value] = result
	}
	return result
}
