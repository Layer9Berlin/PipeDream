package pipeline

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	customstrings "github.com/Layer9Berlin/pipedream/src/custom/strings"
	"github.com/Layer9Berlin/pipedream/src/datastream"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/logrusorgru/aurora/v3"
	"strings"
	"sync"
)

// Run contains everything needed to actually execute the invocation of a pipe
//
// The middleware operates on these objects, triggering further runs or shell invocations
// there are multiple steps to this process:
// 	1. Setup
//		In the setup phase the arguments, connections between inputs and outputs, etc. of each run are defined.
//	2. Closing
//		After the setup, Close() is called to prevent any further changes to input/output connections.
//  3. Execution
//		The shell command is executed and data is piped through the defined input/output connections. Note that some
//		middleware might start additional runs in the execution phase. For example, the `when` middleware for
//		conditional execution will trigger runs based on whether the result of previous runs satisfies a certain condition
//  4. Completion
//  5. Log

type Run struct {
	// arguments are a mix of definition arguments, invocation arguments and changes made by middleware
	arguments map[string]interface{}
	// Identifier is a unique name for pipeline to be executed
	//
	// Note that anonymous pipes without an identifier can have invocation arguments, but no definition
	Identifier *string
	Id         string
	// Definition references the definition matching the pipeline identifier, if any
	Definition *Definition
	// InvocationArguments are passed to the pipe at the time of invocation / run creation
	InvocationArguments map[string]interface{}

	argumentsMutex *sync.RWMutex

	// Stdin is a data stream through which the run's input is passed
	Stdin *datastream.ComposableDataStream
	// Stdout is a data stream through which the run's output is passed
	Stdout *datastream.ComposableDataStream
	// Stderr is a data stream through which the run's stderr output is passed
	Stderr *datastream.ComposableDataStream
	// ExitCode is the exit code of the run's shell command, if any
	ExitCode *int

	// IndefiniteInput indicates whether the run accepts user input from the OS
	//
	// Runs with indefinite input cannot wait for the stdin to complete, for obvious reasons...
	// However, they do close their input once the shell command has finished.
	IndefiniteInput bool

	// Log is the dedicated logger for this run
	//
	// We need to organize our logs by run, so that the order of entries remains consistent
	// during parallel execution of several pipelines.
	Log *Logger

	// Parent is run that started this run, if any
	Parent *Run

	// a run can must be closed exactly once
	// this will close all the data streams
	// and wait for the run to complete
	closed bool

	// after closing, the run will keep executing for a while
	// when everything has been processed, the run will complete
	completed           bool
	completionWaitGroup *sync.WaitGroup

	// WaitGroup will block as long as the run is still in progress
	//
	// It will only unblock when the shell command has completed and all additional tasks have completed
	// For example, a middleware might use the WaitGroup to ensure that the run's Log is still available
	// to be written to even after possible shell command completion
	waitGroup           *sync.WaitGroup
	processingWaitGroup *sync.WaitGroup

	mutex *sync.RWMutex

	StartWaitGroup *sync.WaitGroup

	cancelled   bool
	cancelHooks []func() error
}

// NewRun creates a new Run with the specified identifier, invocation arguments, definition and parent
//
// All passed parameters are optional.
func NewRun(
	identifier *string,
	invocationArguments map[string]interface{},
	definition *Definition,
	parent *Run,
) (*Run, error) {
	arguments := stringmap.CopyMap(invocationArguments)
	if definition != nil {
		err := stringmap.MergeIntoMap(arguments, definition.DefinitionArguments)
		if err != nil {
			return nil, err
		}
	}

	randomUUID, _ := uuid.GenerateUUID()
	run := &Run{
		arguments:  arguments,
		Definition: definition,
		Identifier: identifier,
		Id:         randomUUID,

		ExitCode: nil,

		argumentsMutex: &sync.RWMutex{},

		Parent: parent,

		closed:              false,
		completed:           false,
		completionWaitGroup: &sync.WaitGroup{},
		waitGroup:           &sync.WaitGroup{},
		processingWaitGroup: &sync.WaitGroup{},

		StartWaitGroup: &sync.WaitGroup{},

		mutex: &sync.RWMutex{},

		cancelled:   false,
		cancelHooks: make([]func() error, 0, 10),
	}

	if parent == nil {
		run.Log = NewLogger(run, 0)
	} else {
		run.Log = NewLogger(run, parent.Log.Indentation+2)
		run.Log.SetLevel(parent.Log.Level())
	}

	run.completionWaitGroup.Add(1)
	run.processingWaitGroup.Add(1)

	errorCallback := run.Log.PossibleError

	run.Stdin = datastream.NewComposableDataStream("stdin", errorCallback)
	run.Stdout = datastream.NewComposableDataStream("stdout", errorCallback)
	run.Stderr = datastream.NewComposableDataStream("stderr", errorCallback)

	return run, nil
}

// Close closes the Run's input, output, sterr data streams and log, which is required for execution & completion
//
// It is fine to call Close multiple times. After the first, subsequent calls will have no effect.
func (run *Run) Close() {
	run.mutex.Lock()
	if run.closed {
		run.mutex.Unlock()
		return
	}
	run.closed = true
	run.mutex.Unlock()
	defer run.processingWaitGroup.Done()

	run.Log.Trace(
		fields.Message("closing | "+run.String()),
		fields.Symbol("⏏️"),
		fields.Color("lightgray"),
	)

	run.Stdin.Close()
	run.Stdout.Close()
	run.Stderr.Close()

	run.waitGroup.Add(1)
	go func() {
		run.Stdout.Wait()
		run.Stderr.Wait()
		if !run.IndefiniteInput {
			run.Stdin.Wait()
		}

		run.Log.Info(
			fields.Symbol("✔"),
			fields.Message("completed | "+run.String()),
			fields.Color("green"),
		)

		run.waitGroup.Done()
	}()

	go func() {
		run.waitGroup.Wait()

		run.mutex.Lock()
		run.completed = true
		run.mutex.Unlock()

		run.completionWaitGroup.Done()

		// the log needed to be kept open
		// while new entries might be coming in
		// now that all the data has been processed,
		// we can close the log
		run.Log.Close()
	}()
}

// Wait halts execution until the run has completed
func (run *Run) Wait() {
	run.processingWaitGroup.Wait()
	run.Stdout.Wait()
	run.Stderr.Wait()
	if !run.IndefiniteInput {
		run.Stdin.Wait()
	}
	run.waitGroup.Wait()
	run.completionWaitGroup.Wait()
}

// Completed indicates whether the run has finished executing, logging etc.
func (run *Run) Completed() bool {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	return run.completed
}

// Name returns the run's identifier or "anonymous", if the identifier is nil
func (run *Run) Name() string {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	if run.Identifier == nil {
		return "anonymous"
	}
	return *run.Identifier
}

// String returns a string description of the run suitable for logging
//
// The value will keep changing until the run has completed
func (run *Run) String() string {
	components := make([]string, 0, 10)
	name := run.DisplayString()
	logSummary := run.Log.Summary()
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	components = append(components, fmt.Sprint(aurora.Bold(customstrings.Shorten(name, 128))))
	if run.completed {
		if run.Stdin.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Gray(12, fmt.Sprint("↘️", customstrings.PrettyPrintedByteCount(run.Stdin.Len())))))
		}
		if run.Stdout.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Gray(12, fmt.Sprint("↗️", customstrings.PrettyPrintedByteCount(run.Stdout.Len())))))
		}
		if run.Stderr.Len() > 0 {
			components = append(components, fmt.Sprint(aurora.Red(fmt.Sprint("⛔️", customstrings.PrettyPrintedByteCount(run.Stderr.Len())))))
		}
		if len(logSummary) > 0 {
			components = append(components, logSummary)
		}
	}
	return strings.Join(components, "  ")
}

func (run *Run) GraphLabel() string {
	displayString := run.DisplayString()
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	if run.Log != nil && run.Log.errors != nil && run.Log.errors.Len() > 0 {
		return fmt.Sprintf("✘ %v", displayString)
	}
	if run.cancelled {
		return fmt.Sprintf("⎋ %v", displayString)
	}
	if run.completed {
		return fmt.Sprintf("✔ %v", displayString)
	}
	if run.closed {
		return fmt.Sprintf("↺ %v", displayString)
	}
	return fmt.Sprintf("🔜 %v", displayString)
}

func (run *Run) GraphGroup() string {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	if run.Log != nil && run.Log.errors != nil && run.Log.errors.Len() > 0 {
		return "error"
	}
	if run.cancelled {
		return "cancelled"
	}
	if run.completed {
		return "success"
	}
	if run.closed {
		return "active"
	}
	return "waiting"
}

func (run *Run) DisplayString() string {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	var name *string = nil
	nameArg, err := run.ArgumentAtPath("description")
	if err == nil && nameArg != nil {
		if nameAsString, nameIsString := nameArg.(string); nameIsString && len(nameAsString) > 0 {
			name = &nameAsString
		}
	}
	if (name == nil || *name == "") && run.Identifier != nil {
		prettyName := customstrings.IdentifierToDisplayName(*run.Identifier)
		name = &prettyName
	}
	if name == nil || *name == "" {
		return "anonymous"
	}
	return *name
}

// ArgumentsCopy is a deep copy of the run's arguments that can be safely mutated
func (run *Run) ArgumentsCopy() map[string]interface{} {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return stringmap.CopyMap(run.arguments)
}

// ArgumentAtPath returns the value of the run's arguments at the specified path
func (run *Run) ArgumentAtPath(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return stringmap.GetValueInMap(run.arguments, path...)
}

// HaveArgumentAtPath indicates whether the run's arguments contain a value at the specified path
func (run *Run) HaveArgumentAtPath(path ...string) bool {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return stringmap.HaveValueInMap(run.arguments, path...)
}

// ArgumentAtPathIncludingParents looks up the argument path within the run's arguments or, failing that, its parents
//
// ArgumentAtPathIncludingParents will keep traversing parents until a value is found
func (run *Run) ArgumentAtPathIncludingParents(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	value, err := stringmap.GetValueInMap(run.arguments, path...)
	if err == nil {
		return value, nil
	}

	if run.Parent == nil {
		return nil, fmt.Errorf("value does not exist at path")
	}
	return run.Parent.ArgumentAtPathIncludingParents(path...)
}

// SetArguments overwrites the run's arguments entirely
func (run *Run) SetArguments(value map[string]interface{}) {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	run.arguments = value
}

// SetArgumentAtPath overwrites the run's argument at the specified path, creating additional levels of nesting if required
func (run *Run) SetArgumentAtPath(value interface{}, path ...string) error {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	err := stringmap.SetValueInMap(run.arguments, value, path...)
	return err
}

// RemoveArgumentAtPath removes the run's argument at the specified path
func (run *Run) RemoveArgumentAtPath(path ...string) error {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	err := stringmap.RemoveValueInMap(run.arguments, path...)
	return err
}

// AddCancelHook adds a hook that will be executed when the run is cancelled
//
// Use this to implement cancel functionality in middleware.
func (run *Run) AddCancelHook(cancelHook func() error) {
	run.mutex.Lock()
	defer run.mutex.Unlock()
	run.cancelHooks = append(run.cancelHooks, cancelHook)
}

// Cancel cancels the run without waiting for execution to complete
func (run *Run) Cancel() error {
	run.mutex.Lock()
	defer run.mutex.Unlock()
	run.cancelled = true
	var err = &multierror.Error{}
	for _, cancelHook := range run.cancelHooks {
		err = multierror.Append(err, cancelHook())
	}
	return err.ErrorOrNil()
}

// Cancelled indicates whether the run has been cancelled
func (run *Run) Cancelled() bool {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	return run.cancelled
}

func (run *Run) DontCompleteBefore(executionFunction func()) {
	run.mutex.RLock()
	defer run.mutex.RUnlock()
	if run.closed {
		panic("trying to register a completion waiter on a run that has already been closed")
	}

	run.waitGroup.Add(1)
	go func() {
		defer func() {
			run.waitGroup.Done()
		}()
		executionFunction()
	}()
}
