// Package parsing provides a parser for pipeline yaml files
package parsing

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
)

type Parser struct {
	readFile              func(filename string) ([]byte, error)
	findByGlob            func(pattern string) ([]string, error)
	RecursivelyAddImports func(paths []string) ([]string, error)
}

func NewParser(options ...ParserOption) *Parser {
	parser := &Parser{
		readFile:   ioutil.ReadFile,
		findByGlob: filepath.Glob,
	}
	parser.RecursivelyAddImports = func(filePaths []string) ([]string, error) {
		return parser.recursiveImportStep(filePaths, []string{})
	}
	for _, applyOption := range options {
		applyOption(parser)
	}
	return parser
}

func (parser *Parser) ParsePipelineFiles(allPipelineFilePaths []string, builtIn bool) (
	defaults pipeline.DefaultSettings,
	definitions pipeline.PipelineDefinitionsLookup,
	files []pipeline.File,
	returnErr error,
) {
	// the parsed definitions have a slightly different format
	// this allows the yaml to be concise
	// GO types are not quite flexible enough for this
	definitions = pipeline.PipelineDefinitionsLookup{}
	files = make([]pipeline.File, 0, len(allPipelineFilePaths))
	// TODO: take care of hooks
	for index, pipelineFilePath := range allPipelineFilePaths {
		fileData, err := parser.readFile(pipelineFilePath)
		if err != nil {
			returnErr = err
			return
		}
		pipelineFile := pipeline.File{
			FileName: filepath.Base(pipelineFilePath),
		}
		err = yaml.Unmarshal(fileData, &pipelineFile)
		if err != nil {
			returnErr = fmt.Errorf("unable to parse file %q: %w", pipelineFilePath, err)
			return
		}

		newDefinitions := parser.ProcessPipelineFile(pipelineFile, builtIn)

		for pipelineKey, pipelineDefinition := range newDefinitions {
			if existingParsedDefinitions, ok := definitions[pipelineKey]; ok {
				definitions[pipelineKey] = append(existingParsedDefinitions, pipelineDefinition...)
			} else {
				definitions[pipelineKey] = pipelineDefinition
			}
		}

		if index == 0 && !builtIn {
			defaults = pipelineFile.Default
		}

		files = append(files, pipelineFile)
	}
	return
}

func (parser *Parser) ProcessPipelineFile(
	pipelineFile pipeline.File,
	builtIn bool,
) pipeline.PipelineDefinitionsLookup {
	// iterate through the pipelines and create a new definition for each
	pipelineDefinitions := pipeline.PipelineDefinitionsLookup{}
	for pipelineKey, pipelineValues := range pipelineFile.Public {
		pipelineDefinition := pipeline.NewPipelineDefinition(pipelineValues, pipelineFile.FileName, true, builtIn)
		pipelineDefinitions[pipelineKey] = []pipeline.PipelineDefinition{*pipelineDefinition}
	}
	for pipelineKey, pipelineValues := range pipelineFile.Private {
		pipelineDefinition := pipeline.NewPipelineDefinition(pipelineValues, pipelineFile.FileName, false, builtIn)
		if existingPipelineDefinition, ok := pipelineDefinitions[pipelineKey]; ok {
			pipelineDefinitions[pipelineKey] = append(existingPipelineDefinition, *pipelineDefinition)
		} else {
			pipelineDefinitions[pipelineKey] = []pipeline.PipelineDefinition{*pipelineDefinition}
		}
	}
	return pipelineDefinitions
}
