// Package selectmiddleware provides a middleware that shows a selection prompt to the user
package selectmiddleware

import (
	"bytes"
	customio "github.com/Layer9Berlin/pipedream/src/custom/io"
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
	"io/ioutil"
	"os"
)

// Middleware is a user selection handler
type Middleware struct {
	osStdin  io.ReadCloser
	osStdout io.WriteCloser
}

// String is a human-readable description
func (selectMiddleware Middleware) String() string {
	return "select"
}

// NewMiddleware creates a new middleware instance
func NewMiddleware() Middleware {
	return NewMiddlewareWithStdinAndStdout(os.Stdin, os.Stdout)
}

// NewMiddlewareWithStdinAndStdout creates a new middleware instance with the specified stdin and stdout
func NewMiddlewareWithStdinAndStdout(stdin io.ReadCloser, stdout io.WriteCloser) Middleware {
	return Middleware{
		osStdin:  stdin,
		osStdout: customio.NewBellSkipper(stdout),
	}
}

type middlewareArguments struct {
	Initial int
	Options []pipeline.Reference
	Prompt  *string
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (selectMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	executionContext *middleware.ExecutionContext,
) {
	middlewareArguments := middlewareArguments{
		Initial: 0,
		Options: make([]pipeline.Reference, 0, 16),
	}
	pipeline.ParseArguments(&middlewareArguments, "select", run)

	next(run)

	if len(middlewareArguments.Options) > 0 {
		label := "Please select an option"
		if middlewareArguments.Prompt != nil {
			label = *middlewareArguments.Prompt
		}
		items := make([]string, 0, len(middlewareArguments.Options))
		for _, referenceOption := range middlewareArguments.Options {
			for pipelineIdentifier, pipelineArguments := range referenceOption {
				identifier := "-"
				if description, ok := pipelineArguments["description"]; ok {
					if descriptionString, ok := description.(string); ok {
						identifier = descriptionString
					}
				}
				if identifier == "-" {
					if pipelineIdentifier != nil {
						identifier = *pipelineIdentifier
					}
				}
				items = append(items, identifier)
			}
		}

		stdinCopy := run.Stdin.Copy()
		stdoutWriter := run.Stdout.WriteCloser()
		stderrWriter := run.Stderr.WriteCloser()
		go func() {
			completeStdin, _ := ioutil.ReadAll(stdinCopy)
			if len(completeStdin) > 0 {
				_, _ = selectMiddleware.osStdout.Write(completeStdin)
			}

			selectionIndex, _, err := executionContext.UserPromptImplementation(
				label,
				items,
				middlewareArguments.Initial,
				5,
				selectMiddleware.osStdin,
				selectMiddleware.osStdout,
			)
			run.Log.PossibleError(err)

			selectedPipelineReference := middlewareArguments.Options[selectionIndex]
			selectedPipelineIdentifier := ""
			selectedPipelineArguments := make(map[string]interface{}, 12)
			for pipelineIdentifier, pipelineArguments := range selectedPipelineReference {
				if pipelineIdentifier != nil {
					selectedPipelineIdentifier = *pipelineIdentifier
					selectedPipelineArguments = pipelineArguments
				}
			}
			run.Log.Trace(
				fields.Symbol("👈"),
				fields.Message("user selected pipeline"),
				fields.Info(selectedPipelineIdentifier),
				fields.Middleware(selectMiddleware),
			)
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(&selectedPipelineIdentifier),
				middleware.WithArguments(selectedPipelineArguments),
				middleware.WithSetupFunc(func(childRun *pipeline.Run) {
					run.Log.Trace(
						fields.DataStream(selectMiddleware, "copy parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(bytes.NewReader(completeStdin))
				}),
				middleware.WithTearDownFunc(func(childRun *pipeline.Run) {
					run.Log.Trace(
						fields.DataStream(selectMiddleware, "copy child stdout into parent stdout")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutWriter)
					run.Log.Trace(
						fields.DataStream(selectMiddleware, "copy child stderr into parent stderr")...,
					)
					childRun.Stderr.StartCopyingInto(stderrWriter)
					go func() {
						childRun.Wait()
						_ = stdoutWriter.Close()
						_ = stderrWriter.Close()
					}()
				}))
		}()
		next(run)
		return
	}
}
