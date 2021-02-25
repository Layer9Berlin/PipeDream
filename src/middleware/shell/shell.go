// Package shell provides a middleware to execute commands in a shell
package shell

import (
	"errors"
	"fmt"
	customstrings "github.com/Layer9Berlin/pipedream/src/custom/strings"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Middleware is a shell command runner
type Middleware struct {
	ExecutorCreator func() commandExecutor
	osStdin         io.Reader
	osStdout        io.ReadWriter
	osStderr        io.ReadWriter
}

// String is a human-readable description
func (shellMiddleware Middleware) String() string {
	return "shell"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return NewMiddlewareWithExecutorCreator(func() commandExecutor { return newDefaultCommandExecutor() })
}

// NewMiddlewareWithExecutorCreator creates a new middleware instance with the specified executor creation function
func NewMiddlewareWithExecutorCreator(executorCreator func() commandExecutor) Middleware {
	return Middleware{
		ExecutorCreator: executorCreator,
		osStdin:         os.Stdin,
		osStdout:        os.Stdout,
		osStderr:        os.Stderr,
	}
}

type middlewareArguments struct {
	Args        []interface{}
	Dir         *string
	Exec        string
	Indefinite  bool
	Interactive bool
	Login       bool
	Run         *string
	Quote       string
}

func newMiddlewareArguments() middlewareArguments {
	return middlewareArguments{
		Args:        make([]interface{}, 0, 10),
		Exec:        "sh",
		Indefinite:  false,
		Interactive: false,
		Login:       false,
		Quote:       "double",
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (shellMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(pipelineRun *pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	arguments := newMiddlewareArguments()
	pipeline.ParseArguments(&arguments, "shell", run)

	if arguments.Run == nil {
		next(run)
	} else {
		shellArgumentsAsArray, err := shellCommandArguments(arguments)
		if err != nil {
			run.Log.Error(
				err,
				fields.Middleware(shellMiddleware),
			)
			return
		}

		if len(shellArgumentsAsArray) > 0 {
			*arguments.Run = fmt.Sprintf("%v %v", *arguments.Run, strings.Join(shellArgumentsAsArray, " "))
		}

		if arguments.Dir != nil {
			*arguments.Run = fmt.Sprintf("cd %v && %v", *arguments.Dir, *arguments.Run)
		}

		if arguments.Indefinite || arguments.Interactive {
			run.IndefiniteInput = true
		}

		// allow other middleware to make changes now that we have set the directory
		// this is handy for things like the docker and the SSH middleware
		// since we want to change directories on the remote service instead of locally
		next(run)

		executor := shellMiddleware.ExecutorCreator()

		commandComponents := make([]string, 0, 12)
		if arguments.Login {
			commandComponents = append(commandComponents, "-l")
		}
		commandComponents = append(commandComponents, []string{"-c", *arguments.Run}...)

		executor.Init(arguments.Exec, commandComponents...)

		cmdStdin := executor.CmdStdin()
		var stdinIntercept io.ReadWriteCloser = nil
		if arguments.Interactive {
			// in interactive mode, we want to be ready to read user input
			// and show all output in the console
			stdinIntercept = run.Stdin.Intercept()
			multiWriter := io.MultiWriter(cmdStdin, stdinIntercept)
			go func() {
				for {
					_, err = io.Copy(multiWriter, stdinIntercept)
					if err != nil {
						break
					}
					time.Sleep(200)
				}
			}()
			go func() {
				for {
					_, err = io.Copy(multiWriter, shellMiddleware.osStdin)
					if err != nil {
						break
					}
					time.Sleep(200)
				}
			}()

			// the shell command's stdout should be added to the pipe's output
			// and the combination should be copied to the console
			cmdStdout := executor.CmdStdout()
			run.Stdout.MergeWith(cmdStdout)
			run.Stdout.StartCopyingInto(shellMiddleware.osStdout)

			// same with stderr to preserve the behaviour that both stdout and stderr
			// are shown in the terminal
			cmdStderr := executor.CmdStderr()
			run.Stderr.MergeWith(cmdStderr)
			run.Stderr.StartCopyingInto(shellMiddleware.osStderr)
		} else {
			run.Stdin.StartCopyingInto(cmdStdin)
			run.Stdout.MergeWith(executor.CmdStdout())
			run.Stderr.MergeWith(executor.CmdStderr())
		}

		run.Log.Debug(
			fields.Symbol(">_"),
			fields.Message(executor.String()),
			fields.Middleware(shellMiddleware),
		)

		go func() {
			run.Log.PossibleError(executor.Start())
		}()

		run.AddCancelHook(func() error {
			run.Log.Warn(
				fields.Symbol("⎋"),
				fields.Message("cancelled"),
				fields.Info(executor.String()),
			)
			return executor.Kill()
		})

		if !run.IndefiniteInput {
			go func() {
				run.Stdin.Wait()
				run.Log.PossibleError(cmdStdin.Close())
			}()
		}

		run.WaitGroup.Add(1)
		go func() {
			defer run.WaitGroup.Done()
			if !run.IndefiniteInput {
				run.Stdin.Wait()
			}
			run.Stdout.Wait()
			run.Stderr.Wait()
			err := executor.Wait()
			// reset so that later cancellation will not result in error
			executor.Clear()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode := exitErr.ExitCode()
					run.ExitCode = &exitCode
					run.Log.Error(fmt.Errorf("command exited with non-zero exit code: %w", exitErr))
				} else {
					run.Log.Error(err)
				}
			} else {
				exitCode := 0
				run.ExitCode = &exitCode
			}
			if stdinIntercept != nil {
				err = stdinIntercept.Close()
				run.Log.PossibleError(err)
			}
		}()
	}
}

func shellCommandArguments(pipeArguments middlewareArguments) ([]string, error) {
	middlewareArguments := make([]string, 0, 10)
	for _, argumentItem := range pipeArguments.Args {
		switch typedArgument := argumentItem.(type) {
		case string:
			middlewareArguments = append(middlewareArguments, typedArgument)
		case map[string]interface{}:
			middlewareArguments = append(middlewareArguments, convertMapToStringArray(typedArgument, pipeArguments.Quote)...)
		default:
			return nil, errors.New("provided arguments are in an unknown format - please provide either a map or a (mixed type) array of maps and string")
		}
	}
	return middlewareArguments, nil
}

func convertMapToStringArray(values map[string]interface{}, quoteType string) []string {
	stringResults := make([]string, 0, len(values))
	for argumentKey, argumentValue := range values {
		if argumentValueAsString, argumentValueIsString := argumentValue.(string); argumentValueIsString {
			if strings.HasPrefix(argumentKey, "-") {
				stringResults = append(stringResults, fmt.Sprintf("%v%v", argumentKey, customstrings.QuoteValue(argumentValueAsString, quoteType)))
			} else if len(argumentKey) == 1 {
				stringResults = append(stringResults, fmt.Sprintf("-%v %v", argumentKey, customstrings.QuoteValue(argumentValueAsString, quoteType)))
			} else {
				stringResults = append(stringResults, fmt.Sprintf("--%v=%v", argumentKey, customstrings.QuoteValue(argumentValueAsString, quoteType)))
			}
		} else if argumentValueAsBool, argumentValueIsBool := argumentValue.(bool); argumentValueIsBool {
			if argumentValueAsBool {
				if strings.HasPrefix(argumentKey, "-") {
					stringResults = append(stringResults, argumentKey)
				} else if len(argumentKey) == 1 {
					stringResults = append(stringResults, fmt.Sprintf("-%v", argumentKey))
				} else {
					stringResults = append(stringResults, fmt.Sprintf("--%v", argumentKey))
				}
			}
		}
	}
	// sort alphabetically for predictable results
	sort.StringSlice.Sort(stringResults)
	return stringResults
}
