package models

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"pipedream/src/helpers/custom_strings"
	"pipedream/src/helpers/string_map"
	"pipedream/src/logging"
	"pipedream/src/logging/log_fields"
	"strings"
	"sync"
)

// a pipeline run contains everything needed to actually execute the invocation of a pipe
// the middleware operates on these objects, triggering further runs or shell invocations
// there are three steps to this process:
// 	1. Setup
//		In the setup phase the arguments, connections between inputs and outputs, etc. of each run are defined.
//	2. Finalization
//		After the setup, finalize() is called to prevent any further changes to input/output connections.
//  3. Execution
//		The shell command is executed and data is piped through the defined input/output connections. Note that some
//		middleware might start additional runs in the execution phase. For example, the `when` middleware for
//		conditional execution will trigger runs based on whether the result of previous runs satisfies a certain condition
type PipelineRun struct {
	// the stored arguments are a mix of definition arguments, invocation arguments and changes made by middleware
	arguments map[string]interface{}
	// if the identifier has been defined before, this will be the resolved reference to the definition
	Definition *PipelineDefinition
	// an optional identifier - note that anonymous pipes without identifier can have invocation arguments, but no definition
	Identifier *string
	// the invocation arguments are what is passed to the pipe at the time of invocation / run creation
	InvocationArguments map[string]interface{}

	argumentsMutex *sync.RWMutex

	Stdin    *ComposableDataStream
	Stdout   *ComposableDataStream
	Stderr   *ComposableDataStream
	ExitCode *int

	Log *PipelineRunLogger

	Parent *PipelineRun

	// a run can must be closed exactly once
	// this will close all the data streams
	// and wait for the run to complete
	closeMutex *sync.Mutex
	closed     bool

	// after closing, the run will keep executing for a while
	// when everything has been processed, the run will complete
	completed           bool
	completionWaitGroup *sync.WaitGroup
	LogClosingWaitGroup *sync.WaitGroup

	cancelled   bool
	cancelHooks []func() error

	Synchronous    bool
	StartWaitGroup *sync.WaitGroup
}

func NewPipelineRun(
	identifier *string,
	invocationArguments map[string]interface{},
	definition *PipelineDefinition,
	parent *PipelineRun,
) (*PipelineRun, error) {
	arguments := string_map.CopyMap(invocationArguments)
	if definition != nil {
		err := string_map.MergeIntoMap(arguments, definition.DefinitionArguments)
		if err != nil {
			return nil, err
		}
	}

	run := &PipelineRun{
		arguments:  arguments,
		Definition: definition,
		Identifier: identifier,

		ExitCode: nil,

		argumentsMutex: &sync.RWMutex{},

		Parent: parent,

		closeMutex:          &sync.Mutex{},
		closed:              false,
		completed:           false,
		completionWaitGroup: &sync.WaitGroup{},
		LogClosingWaitGroup: &sync.WaitGroup{},

		cancelled:   false,
		cancelHooks: make([]func() error, 0, 10),

		Synchronous:    false,
		StartWaitGroup: &sync.WaitGroup{},
	}

	if parent == nil {
		run.Log = NewPipelineRunLogger(run, 0)
	} else {
		run.Log = NewPipelineRunLogger(run, parent.Log.Indentation+2)
		run.Log.SetLevel(parent.Log.Level())
	}

	errorCallback := func(err error) {
		run.Log.PossibleError(err)
	}
	run.Stdin = NewComposableDataStream("stdin", errorCallback)
	run.Stdout = NewComposableDataStream("stdout", errorCallback)
	run.Stderr = NewComposableDataStream("stderr", errorCallback)

	return run, nil
}

func (run *PipelineRun) Close() {
	run.closeMutex.Lock()
	defer run.closeMutex.Unlock()
	if run.closed {
		return
	}
	run.closed = true

	run.Log.TraceWithFields(
		log_fields.Message("closing | "+run.String()),
		log_fields.Symbol("⏏️"),
		log_fields.Color("lightgray"),
	)

	run.Stdin.Close()
	run.Stdout.Close()
	run.Stderr.Close()

	run.completionWaitGroup.Add(1)
	go func() {
		defer run.completionWaitGroup.Done()

		run.Stdout.Wait()
		run.Stderr.Wait()
		run.Stdin.Wait()

		run.completed = true

		run.LogClosingWaitGroup.Wait()

		run.Log.DebugWithFields(
			log_fields.Symbol("✔"),
			log_fields.Message("completed | "+run.String()),
			log_fields.Color("green"),
		)

		// the log needed to be kept open
		// while new entries might be coming in
		// now that all the data has been processed,
		// we can close the log
		run.Log.Close()
		run.Log.Wait()
	}()
}

func (run *PipelineRun) Wait() {
	run.Stdout.Wait()
	run.Stderr.Wait()
	run.Stdin.Wait()
	run.Log.Wait()
	// ensure that run.Completed() returns true after the call to run.Wait()
	run.completionWaitGroup.Wait()
}

func (run *PipelineRun) Completed() bool {
	return run.completed
}

func (run *PipelineRun) String() string {
	components := make([]string, 0, 10)
	var name *string = nil
	nameArg, err := run.ArgumentAtPath("description")
	if err == nil {
		if nameAsString, nameIsString := nameArg.(string); nameIsString && len(nameAsString) > 0 {
			name = &nameAsString
		}
	}
	if (name == nil || *name == "") && run.Identifier != nil {
		prettyName := custom_strings.IdentifierToDisplayName(*run.Identifier)
		name = &prettyName
	}
	if name == nil || *name == "" {
		anonymousName := "anonymous"
		name = &anonymousName
	}
	components = append(components, fmt.Sprint(aurora.Bold(logging.ShortenString(*name, 128))))
	if run.completed {
		if run.Stdin.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Gray(12, fmt.Sprint("↘️", custom_strings.PrettyPrintedByteCount(run.Stdin.Len())))))
		}
		if run.Stdout.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Gray(12, fmt.Sprint("↗️", custom_strings.PrettyPrintedByteCount(run.Stdout.Len())))))
		}
		if run.Stderr.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Red(fmt.Sprint("⛔️", custom_strings.PrettyPrintedByteCount(run.Stderr.Len())))))
		}
		logSummary := run.Log.Summary()
		if len(logSummary) > 0 {
			components = append(components, logSummary)
		}
	}
	return strings.Join(components, "  ")
}

func (run *PipelineRun) ArgumentsCopy() map[string]interface{} {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return string_map.CopyMap(run.arguments)
}

func (run *PipelineRun) ArgumentAtPath(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return string_map.GetValueInMap(run.arguments, path...)
}

func (run *PipelineRun) ArgumentAtPathIncludingParents(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	value, err := string_map.GetValueInMap(run.arguments, path...)
	if err == nil {
		return value, nil
	}

	if run.Parent == nil {
		return nil, fmt.Errorf("value does not exist at path")
	} else {
		return run.Parent.ArgumentAtPathIncludingParents(path...)
	}
}

func (run *PipelineRun) SetArguments(value map[string]interface{}) {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	run.arguments = value
}

func (run *PipelineRun) SetArgumentAtPath(value interface{}, path ...string) error {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	err := string_map.SetValueInMap(run.arguments, value, path...)
	return err
}

func (run *PipelineRun) AddCancelHook(cancelHook func() error) {
	run.cancelHooks = append(run.cancelHooks, cancelHook)
}

func (run *PipelineRun) Cancel() error {
	run.cancelled = true
	var err = &multierror.Error{}
	for _, cancelHook := range run.cancelHooks {
		err = multierror.Append(err, cancelHook())
	}
	return err.ErrorOrNil()
}

func (run *PipelineRun) Cancelled() bool {
	return run.cancelled
}
