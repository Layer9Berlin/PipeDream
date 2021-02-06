// Provides a model representing a pipeline run
package pipeline

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	customstrings "github.com/Layer9Berlin/pipedream/src/custom/strings"
	"github.com/Layer9Berlin/pipedream/src/datastream"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"strings"
	"sync"
)

// A pipeline run contains everything needed to actually execute the invocation of a pipe
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
type Run struct {
	// the stored arguments are a mix of definition arguments, invocation arguments and changes made by middleware
	arguments map[string]interface{}
	// if the identifier has been defined before, this will be the resolved reference to the definition
	Definition *PipelineDefinition
	// an optional identifier - note that anonymous pipes without identifier can have invocation arguments, but no definition
	Identifier *string
	// the invocation arguments are what is passed to the pipe at the time of invocation / run creation
	InvocationArguments map[string]interface{}

	argumentsMutex *sync.RWMutex

	Stdin    *datastream.ComposableDataStream
	Stdout   *datastream.ComposableDataStream
	Stderr   *datastream.ComposableDataStream
	ExitCode *int

	Log *PipelineRunLogger

	Parent *Run

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
	parent *Run,
) (*Run, error) {
	arguments := stringmap.CopyMap(invocationArguments)
	if definition != nil {
		err := stringmap.MergeIntoMap(arguments, definition.DefinitionArguments)
		if err != nil {
			return nil, err
		}
	}

	run := &Run{
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
	run.Stdin = datastream.NewComposableDataStream("stdin", errorCallback)
	run.Stdout = datastream.NewComposableDataStream("stdout", errorCallback)
	run.Stderr = datastream.NewComposableDataStream("stderr", errorCallback)

	return run, nil
}

func (run *Run) Close() {
	run.closeMutex.Lock()
	defer run.closeMutex.Unlock()
	if run.closed {
		return
	}
	run.closed = true

	run.Log.TraceWithFields(
		fields.Message("closing | "+run.String()),
		fields.Symbol("⏏️"),
		fields.Color("lightgray"),
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
			fields.Symbol("✔"),
			fields.Message("completed | "+run.String()),
			fields.Color("green"),
		)

		// the log needed to be kept open
		// while new entries might be coming in
		// now that all the data has been processed,
		// we can close the log
		run.Log.Close()
		run.Log.Wait()
	}()
}

func (run *Run) Wait() {
	run.Stdout.Wait()
	run.Stderr.Wait()
	run.Stdin.Wait()
	run.Log.Wait()
	// ensure that run.Completed() returns true after the call to run.Wait()
	run.completionWaitGroup.Wait()
}

func (run *Run) Completed() bool {
	return run.completed
}

func (run *Run) String() string {
	components := make([]string, 0, 10)
	var name *string = nil
	nameArg, err := run.ArgumentAtPath("description")
	if err == nil {
		if nameAsString, nameIsString := nameArg.(string); nameIsString && len(nameAsString) > 0 {
			name = &nameAsString
		}
	}
	if (name == nil || *name == "") && run.Identifier != nil {
		prettyName := customstrings.IdentifierToDisplayName(*run.Identifier)
		name = &prettyName
	}
	if name == nil || *name == "" {
		anonymousName := "anonymous"
		name = &anonymousName
	}
	components = append(components, fmt.Sprint(aurora.Bold(logging.ShortenString(*name, 128))))
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
		logSummary := run.Log.Summary()
		if len(logSummary) > 0 {
			components = append(components, logSummary)
		}
	}
	return strings.Join(components, "  ")
}

func (run *Run) ArgumentsCopy() map[string]interface{} {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return stringmap.CopyMap(run.arguments)
}

func (run *Run) ArgumentAtPath(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	return stringmap.GetValueInMap(run.arguments, path...)
}

func (run *Run) ArgumentAtPathIncludingParents(path ...string) (interface{}, error) {
	run.argumentsMutex.RLock()
	defer run.argumentsMutex.RUnlock()
	value, err := stringmap.GetValueInMap(run.arguments, path...)
	if err == nil {
		return value, nil
	}

	if run.Parent == nil {
		return nil, fmt.Errorf("value does not exist at path")
	} else {
		return run.Parent.ArgumentAtPathIncludingParents(path...)
	}
}

func (run *Run) SetArguments(value map[string]interface{}) {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	run.arguments = value
}

func (run *Run) SetArgumentAtPath(value interface{}, path ...string) error {
	run.argumentsMutex.Lock()
	defer run.argumentsMutex.Unlock()
	err := stringmap.SetValueInMap(run.arguments, value, path...)
	return err
}

func (run *Run) AddCancelHook(cancelHook func() error) {
	run.cancelHooks = append(run.cancelHooks, cancelHook)
}

func (run *Run) Cancel() error {
	run.cancelled = true
	var err = &multierror.Error{}
	for _, cancelHook := range run.cancelHooks {
		err = multierror.Append(err, cancelHook())
	}
	return err.ErrorOrNil()
}

func (run *Run) Cancelled() bool {
	return run.cancelled
}
