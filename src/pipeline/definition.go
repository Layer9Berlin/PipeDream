package pipeline

// HookDefinitions are not currently used
type HookDefinitions = map[string][]string

// DefinitionsLookup maps identifiers to their definitions
type DefinitionsLookup = map[string][]Definition

// Definition defines default arguments for a pipeline identifier
type Definition struct {
	BuiltIn             bool
	DefinitionArguments map[string]interface{}
	FileName            string
	Public              bool
}

// NewDefinition creates a new Definition
func NewDefinition(
	arguments map[string]interface{},
	pipelineFileName string,
	isPublic bool,
	isBuiltIn bool,
) *Definition {
	return &Definition{
		BuiltIn:             isBuiltIn,
		DefinitionArguments: arguments,
		FileName:            pipelineFileName,
		Public:              isPublic,
	}
}

// MergePipelineDefinitions merges two definition lookups into a single one
func MergePipelineDefinitions(definition1 DefinitionsLookup, definition2 DefinitionsLookup) DefinitionsLookup {
	result := DefinitionsLookup{}
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
