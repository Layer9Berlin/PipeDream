package parsers

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"pipedream/src/models"
)

type ParserOption func(parser *Parser)

type Parser struct {
	readFile              func(filename string) ([]byte, error)
	findByGlob            func(pattern string) ([]string, error)
	RecursivelyAddImports func(paths []string) ([]string, error)
}

func WithReadFileImplementation(readFile func(filename string) ([]byte, error)) ParserOption {
	return func(parser *Parser) {
		parser.readFile = readFile
	}
}

func WithFindByGlobImplementation(findByGlob func(pattern string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.findByGlob = findByGlob
	}
}

func WithRecursivelyAddImportsImplementation(recursivelyAddImports func(paths []string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.RecursivelyAddImports = recursivelyAddImports
	}
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
	defaults models.DefaultSettings,
	definitions models.PipelineDefinitionsLookup,
	files []models.PipelineFile,
	returnErr error,
) {
	// the parsed definitions have a slightly different format
	// this allows the yaml to be concise
	// GO types are not quite flexible enough for this
	definitions = models.PipelineDefinitionsLookup{}
	files = make([]models.PipelineFile, 0, len(allPipelineFilePaths))
	// TODO: take care of hooks
	for index, pipelineFilePath := range allPipelineFilePaths {
		fileData, err := parser.readFile(pipelineFilePath)
		if err != nil {
			returnErr = err
			return
		}
		pipelineFile := models.PipelineFile{
			FileName: filepath.Base(pipelineFilePath),
		}
		err = yaml.Unmarshal(fileData, &pipelineFile)
		if err != nil {
			returnErr = err
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
	pipelineFile models.PipelineFile,
	builtIn bool,
) models.PipelineDefinitionsLookup {
	// iterate through the pipelines and create a new definition for each
	pipelineDefinitions := models.PipelineDefinitionsLookup{}
	for pipelineKey, pipelineValues := range pipelineFile.Public {
		definition := models.NewPipelineDefinition(pipelineValues, pipelineFile, true, builtIn)
		pipelineDefinitions[pipelineKey] = []models.PipelineDefinition{*definition}
	}
	for pipelineKey, pipelineValues := range pipelineFile.Private {
		definition := models.NewPipelineDefinition(pipelineValues, pipelineFile, false, builtIn)
		if existingPipelineDefinition, ok := pipelineDefinitions[pipelineKey]; ok {
			pipelineDefinitions[pipelineKey] = append(existingPipelineDefinition, *definition)
		} else {
			pipelineDefinitions[pipelineKey] = []models.PipelineDefinition{*definition}
		}
	}
	return pipelineDefinitions
}
