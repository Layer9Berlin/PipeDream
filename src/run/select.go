package run

import (
	"errors"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/strings"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"io"
	sort "sort"
)

func letUserSelectPipelineFile(executionContext *middleware.ExecutionContext, selectionWindowSize int, input io.ReadCloser, output io.WriteCloser) (*pipeline.File, error) {
	if FileFlag != "" {
		return executionContext.PipelineFileAtPath(FileFlag)
	}

	pipelineFiles := executionContext.SelectableFiles
	if pipelineFiles == nil || len(pipelineFiles) == 0 {
		return nil, fmt.Errorf("no pipeline file found, perhaps you are in the wrong directory")
	}
	if len(pipelineFiles) == 1 {
		return executionContext.PipelineFileAtPath(pipelineFiles[0])
	}

	displayNames := make([]string, 0, len(pipelineFiles))
	for _, pipelineFilePath := range pipelineFiles {
		pipelineFile, err := executionContext.PipelineFileAtPath(pipelineFilePath)
		if err == nil {
			displayNames = append(displayNames, strings.IdentifierToDisplayName(pipelineFile.FileName))
		}
	}

	resultIndex, _, err := executionContext.UserPromptImplementation(
		"Select pipeline file",
		displayNames,
		0,
		selectionWindowSize,
		input,
		output,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("prompt failed: %v", err))
	}

	return executionContext.PipelineFileAtPath(pipelineFiles[resultIndex])
}

func letUserSelectPipelineFileAndPipeline(
	executionContext *middleware.ExecutionContext,
	selectionWindowSize int,
	input io.ReadCloser,
	output io.WriteCloser,
) (string, string, error) {
	pipelineFiles := executionContext.PipelineFiles
	if pipelineFiles == nil || len(pipelineFiles) == 0 {
		return "no-pipelines::handle", "", nil
	}

	pipelineFile, err := letUserSelectPipelineFile(executionContext, 10, input, output)
	if err != nil {
		return "", "", err
	}
	executionContext.Hooks = pipelineFile.Hooks

	pipelineIdentifiers := make([]string, 0, len(pipelineFile.Public)+len(pipelineFile.Private))
	for pipelineIdentifier := range pipelineFile.Public {
		pipelineIdentifiers = append(pipelineIdentifiers, pipelineIdentifier)
	}
	sort.StringSlice.Sort(pipelineIdentifiers)

	if PipelineFlag != "" {
		return PipelineFlag, pipelineFile.FileName, nil
	}

	if len(pipelineIdentifiers) == 1 {
		return pipelineIdentifiers[0], pipelineFile.FileName, nil
	}

	pipelineNames := make([]string, 0, len(pipelineIdentifiers))
	initialSelection := 0
	for index, pipelineIdentifier := range pipelineIdentifiers {
		pipelineNames = append(pipelineNames, strings.IdentifierToDisplayName(pipelineIdentifier))
		if pipelineIdentifier == pipelineFile.Default.Command {
			initialSelection = index
		}
	}

	resultIndex, _, err := executionContext.UserPromptImplementation(
		"Select pipeline",
		pipelineNames,
		initialSelection,
		selectionWindowSize,
		input,
		output,
	)

	if err != nil {
		return "", pipelineFile.FileName, errors.New(fmt.Sprintf("prompt failed: %v", err))
	}

	return pipelineIdentifiers[resultIndex], pipelineFile.FileName, nil
}
