package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pipedream/src/helpers/custom_io"
	"pipedream/src/helpers/custom_strings"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"sort"
	"strings"
)

// Shell Command Runner
type ShellMiddleware struct {
	ExecutorCreator func() CommandExecutor
	osStdin         io.Reader
	osStdout        io.ReadWriter
	osStderr        io.ReadWriter
}

func (shellMiddleware ShellMiddleware) String() string {
	return "shell"
}

func NewShellMiddleware() ShellMiddleware {
	return NewShellMiddlewareWithExecutorCreator(func() CommandExecutor { return NewDefaultCommandExecutor() })
}

func NewShellMiddlewareWithExecutorCreator(executorCreator func() CommandExecutor) ShellMiddleware {
	return ShellMiddleware{
		ExecutorCreator: executorCreator,
		osStdin:         os.Stdin,
		osStdout:        os.Stdout,
		osStderr:        os.Stderr,
	}
}

type ShellMiddlewareArguments struct {
	Args        []interface{}
	Dir         *string
	Exec        string
	Indefinite  bool
	Interactive bool
	Login       bool
	Run         *string
	Quote       string
}

func NewShellMiddlewareArguments() ShellMiddlewareArguments {
	return ShellMiddlewareArguments{
		Args:        make([]interface{}, 0, 10),
		Exec:        "sh",
		Indefinite:  false,
		Interactive: false,
		Login:       false,
		Quote:       "double",
	}
}

func (shellMiddleware ShellMiddleware) Apply(
	run *models.PipelineRun,
	next func(pipelineRun *models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := NewShellMiddlewareArguments()
	middleware.ParseArguments(&arguments, "shell", run)

	if arguments.Run == nil {
		next(run)
	} else {
		shellArgumentsAsArray, err := shellCommandArguments(arguments)
		if err != nil {
			run.Log.Error(
				err,
				log_fields.Middleware(shellMiddleware),
			)
			return
		} else {
			if len(shellArgumentsAsArray) > 0 {
				*arguments.Run = fmt.Sprintf("%v %v", *arguments.Run, strings.Join(shellArgumentsAsArray, " "))
			}
		}

		if arguments.Dir != nil {
			*arguments.Run = fmt.Sprintf("cd %v && %v", *arguments.Dir, *arguments.Run)
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
			executionContext.ActivityIndicator.SetVisible(false)

			// in interactive mode, we want to be ready to read user input
			// and show all output in the console
			stdinIntercept = run.Stdin.Intercept()

			go func() {
				_, _ = io.Copy(cmdStdin, stdinIntercept)
			}()
			go func() {
				// need to wrap the osStdin, as it may return EOF for some time until new user input arrives
				_, _ = io.Copy(io.MultiWriter(stdinIntercept, cmdStdin), custom_io.NewContinuousReader(shellMiddleware.osStdin))
				_ = stdinIntercept.Close()
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
			if !run.Synchronous {
				run.Stdin.StartCopyingInto(cmdStdin)
			}
			run.Stdout.MergeWith(executor.CmdStdout())
			run.Stderr.MergeWith(executor.CmdStderr())
		}

		run.Log.DebugWithFields(
			log_fields.Symbol(">_"),
			log_fields.Message(executor.String()),
			log_fields.Middleware(shellMiddleware),
		)

		// TODO: deal with runs that are both synchronous and interactive
		if run.Synchronous {
			go func() {
				run.Stdin.Wait()
				go func() {
					_, err := io.Copy(cmdStdin, bytes.NewReader(run.Stdin.Bytes()))
					run.Log.PossibleError(err)
				}()
				run.Log.PossibleError(executor.Start())
			}()
		} else {
			run.Log.PossibleError(executor.Start())
		}

		run.AddCancelHook(func() error {
			run.Log.WarnWithFields(
				log_fields.Symbol("⎋"),
				log_fields.Message("cancelled"),
				log_fields.Info(executor.String()),
			)
			return executor.Kill()
		})

		if !run.Synchronous && !arguments.Indefinite && !arguments.Interactive {
			go func() {
				run.Stdin.Wait()
				run.Log.PossibleError(cmdStdin.Close())
			}()
		}

		run.LogClosingWaitGroup.Add(1)
		go func() {
			run.Stdin.Wait()
			run.Stdout.Wait()
			run.Stderr.Wait()
			err := executor.Wait()
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
			run.LogClosingWaitGroup.Done()
		}()
	}
}

func shellCommandArguments(pipeArguments ShellMiddlewareArguments) ([]string, error) {
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
				stringResults = append(stringResults, fmt.Sprintf("%v%v", argumentKey, custom_strings.QuoteValue(argumentValueAsString, quoteType)))
			} else if len(argumentKey) == 1 {
				stringResults = append(stringResults, fmt.Sprintf("-%v %v", argumentKey, custom_strings.QuoteValue(argumentValueAsString, quoteType)))
			} else {
				stringResults = append(stringResults, fmt.Sprintf("--%v=%v", argumentKey, custom_strings.QuoteValue(argumentValueAsString, quoteType)))
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

type CommandExecutor interface {
	Init(name string, arg ...string)
	Kill() error
	CmdStdin() io.WriteCloser
	CmdStdout() io.Reader
	CmdStderr() io.Reader
	Start() error
	String() string
	Wait() error
}

type DefaultCommandExecutor struct {
	command *exec.Cmd
	env     []string
	stopped bool
}

func NewDefaultCommandExecutor() *DefaultCommandExecutor {
	return &DefaultCommandExecutor{
		env:     os.Environ(),
		stopped: false,
	}
}

func (executor *DefaultCommandExecutor) Init(name string, arg ...string) {
	executor.command = exec.Command(name, arg...)
	executor.command.Env = executor.env
}

func (executor *DefaultCommandExecutor) Start() error {
	return executor.command.Start()
}

func (executor *DefaultCommandExecutor) CmdStdin() io.WriteCloser {
	stdin, _ := executor.command.StdinPipe()
	return stdin
}

func (executor *DefaultCommandExecutor) CmdStdout() io.Reader {
	stdout, _ := executor.command.StdoutPipe()
	return stdout
}

func (executor *DefaultCommandExecutor) CmdStderr() io.Reader {
	stderr, _ := executor.command.StderrPipe()
	return stderr
}

func (executor *DefaultCommandExecutor) Wait() error {
	return executor.command.Wait()
}

func (executor *DefaultCommandExecutor) Kill() error {
	return executor.command.Process.Kill()
}

func (executor *DefaultCommandExecutor) String() string {
	return executor.command.String()
}
