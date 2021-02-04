package selectMiddleware

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/helpers/custom_io"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
	"io"
	"io/ioutil"
	"os"
)

type SelectMiddleware struct {
	osStdin  io.ReadCloser
	osStdout io.WriteCloser
}

func (selectMiddleware SelectMiddleware) String() string {
	return "select"
}

func NewSelectMiddleware() SelectMiddleware {
	return NewSelectMiddlewareWithStdinAndStdout(os.Stdin, os.Stdout)
}

func NewSelectMiddlewareWithStdinAndStdout(stdin io.ReadCloser, stdout io.WriteCloser) SelectMiddleware {
	return SelectMiddleware{
		osStdin:  stdin,
		osStdout: custom_io.NewBellSkipper(stdout),
	}
}

type selectMiddlewareArguments struct {
	Initial int
	Options []middleware.PipelineReference
	Prompt  *string
}

func (selectMiddleware SelectMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	arguments := selectMiddlewareArguments{
		Initial: 0,
		Options: make([]middleware.PipelineReference, 0, 16),
	}
	middleware.ParseArguments(&arguments, "select", run)

	next(run)

	if len(arguments.Options) > 0 {
		label := "Please select an option"
		if arguments.Prompt != nil {
			label = *arguments.Prompt
		}
		items := make([]string, 0, len(arguments.Options))
		for _, referenceOption := range arguments.Options {
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

		executionContext.ActivityIndicator.SetVisible(false)

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
				arguments.Initial,
				5,
				selectMiddleware.osStdin,
				selectMiddleware.osStdout,
			)
			run.Log.PossibleError(err)

			selectedPipelineReference := arguments.Options[selectionIndex]
			selectedPipelineIdentifier := ""
			selectedPipelineArguments := make(map[string]interface{}, 12)
			for pipelineIdentifier, pipelineArguments := range selectedPipelineReference {
				if pipelineIdentifier != nil {
					selectedPipelineIdentifier = *pipelineIdentifier
					selectedPipelineArguments = pipelineArguments
				}
			}
			run.Log.TraceWithFields(
				log_fields.Symbol("👈"),
				log_fields.Message("user selected pipeline"),
				log_fields.Info(selectedPipelineIdentifier),
				log_fields.Middleware(selectMiddleware),
			)
			executionContext.FullRun(
				middleware.WithParentRun(run),
				middleware.WithIdentifier(&selectedPipelineIdentifier),
				middleware.WithArguments(selectedPipelineArguments),
				middleware.WithSetupFunc(func(childRun *models.PipelineRun) {
					run.Log.TraceWithFields(
						log_fields.DataStream(selectMiddleware, "copy parent stdin into child stdin")...,
					)
					childRun.Stdin.MergeWith(bytes.NewReader(completeStdin))
				}),
				middleware.WithTearDownFunc(func(childRun *models.PipelineRun) {
					run.Log.TraceWithFields(
						log_fields.DataStream(selectMiddleware, "copy child stdout into parent stdout")...,
					)
					childRun.Stdout.StartCopyingInto(stdoutWriter)
					run.Log.TraceWithFields(
						log_fields.DataStream(selectMiddleware, "copy child stderr into parent stderr")...,
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

type SaveMiddlewareEntry struct {
	path []string
	root *models.PipelineRun
}
