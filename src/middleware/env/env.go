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

// Env Var
type envMiddlewareArguments struct {
	Interpolate string
	Save        *string
}

type EnvMiddleware struct {
	Setenv    func(string, string) error
	ExpandEnv func(string) string
}

func (envMiddleware EnvMiddleware) String() string {
	return "env"
}

func NewEnvMiddleware() EnvMiddleware {
	return NewEnvMiddlewareWithProvider(os.Setenv, os.ExpandEnv)
}

func NewEnvMiddlewareWithProvider(
	setenv func(string, string) error,
	expandEnv func(string) string,
) EnvMiddleware {
	return EnvMiddleware{
		Setenv:    setenv,
		ExpandEnv: expandEnv,
	}
}

func (envMiddleware EnvMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := envMiddlewareArguments{
		Interpolate: "shallow",
		Save:        nil,
	}
	pipeline.ParseArguments(&arguments, "env", run)

	interpolator := NewInterpolator(envMiddleware)
	interpolatedArguments := run.ArgumentsCopy()
	switch arguments.Interpolate {
	case "deep":
		structparse.Strings(interpolator, interpolatedArguments)
		run.SetArguments(interpolatedArguments)
	case "none":
	default:
		newArguments := run.ArgumentsCopy()
		for argumentKey, argumentValue := range interpolatedArguments {
			if argumentValueAsString, argumentValueIsString := argumentValue.(string); argumentValueIsString {
				newArguments[argumentKey] = interpolator.ParseString(argumentValueAsString)
			}
		}
		run.SetArguments(newArguments)
	}
	switch len(interpolator.Substitutions) {
	case 0:
	case 1:
		run.Log.DebugWithFields(
			fields.Symbol("💲"),
			fields.Message("made 1 env var substitution"),
			fields.Info(interpolator.Substitutions),
			fields.Middleware(envMiddleware),
		)
	default:
		run.Log.DebugWithFields(
			fields.Symbol("💲"),
			fields.Message(fmt.Sprintf("made %v env var substitutions", len(interpolator.Substitutions))),
			fields.Info(interpolator.Substitutions),
			fields.Middleware(envMiddleware),
		)
	}

	next(run)

	if arguments.Save != nil {
		// to avoid flakiness, we need to defer subsequent executions,
		// as they will usually want to use the env var we are setting here
		run.Synchronous = true

		run.LogClosingWaitGroup.Add(1)
		go func() {
			run.Stdout.Wait()
			err := envMiddleware.Setenv(*arguments.Save, run.Stdout.String())
			run.Log.PossibleError(err)
			run.LogClosingWaitGroup.Done()
		}()
		run.Log.DebugWithFields(
			fields.Symbol("💲"),
			fields.Message(fmt.Sprintf("saving output")),
			fields.Info("$"+*arguments.Save),
			fields.Middleware(envMiddleware),
		)
	}
}

type Interpolator struct {
	Substitutions map[string]interface{}
	ExpandEnv     func(string) string
}

func NewInterpolator(envMiddleware EnvMiddleware) *Interpolator {
	return &Interpolator{
		Substitutions: make(map[string]interface{}, 10),
		ExpandEnv:     envMiddleware.ExpandEnv,
	}
}

func (interpolator *Interpolator) ParseString(value string) interface{} {
	result := interpolator.ExpandEnv(value)
	if result != value {
		interpolator.Substitutions[value] = result
	}
	return result
}
