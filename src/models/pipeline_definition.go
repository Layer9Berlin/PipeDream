package models

type HookDefinitions = map[string][]string
type PipelineDefinitionsLookup = map[string][]PipelineDefinition

type PipelineDefinition struct {
	BuiltIn             bool
	DefinitionArguments map[string]interface{}
	FileName            string
	Public              bool
}

func NewPipelineDefinition(
	arguments map[string]interface{},
	pipelineFile PipelineFile,
	isPublic bool,
	isBuiltIn bool,
) *PipelineDefinition {
	return &PipelineDefinition{
		BuiltIn:             isBuiltIn,
		DefinitionArguments: arguments,
		FileName:            pipelineFile.FileName,
		Public:              isPublic,
	}
}

func MergePipelineDefinitions(definition1 PipelineDefinitionsLookup, definition2 PipelineDefinitionsLookup) PipelineDefinitionsLookup {
	result := PipelineDefinitionsLookup{}
	for key, value := range definition1 {
		result[key] = value
	}
	for key, value := range definition2 {
		if existingParsedDefinitions, ok := result[key]; ok {
			result[key] = append(existingParsedDefinitions, value...)
		} else {
			result[key] = value
		}
	}

	return result
}
