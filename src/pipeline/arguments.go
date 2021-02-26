package pipeline

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/custom/stringmap"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

// Arguments maps the string key corresponding to each middleware's identifier to its specific arguments
type Arguments = map[string]interface{}

// Reference is a map containing a single value indexed by the pipeline's identifier (possibly nil)
type Reference = map[*string]Arguments

var pipelineReferenceType = reflect.TypeOf(Reference{})

func pipelineReferenceDecodeHook(
	_ reflect.Type,
	outputType reflect.Type,
	value interface{},
) (interface{}, error) {
	if outputType == pipelineReferenceType {
		if valueAsString, valueIsString := value.(string); valueIsString {
			return Reference{
				&valueAsString: make(map[string]interface{}, 0),
			}, nil
		}
		if valueAsMap, valueIsMap := value.(map[string]interface{}); valueIsMap {
			if len(valueAsMap) != 1 {
				return nil, fmt.Errorf("invalid pipeline reference: you must provide exactly one key/value combination")
			}
		} else if valueAsPointerMap, valueIsPointerMap := value.(map[*string]interface{}); valueIsPointerMap {
			if len(valueAsPointerMap) != 1 {
				return nil, fmt.Errorf("invalid pipeline reference: you must provide exactly one key/value combination")
			}
		} else if referenceAsInterfaceMap, referenceIsInterfaceMap := value.(map[interface{}]interface{}); referenceIsInterfaceMap {
			if len(referenceAsInterfaceMap) != 1 {
				return nil, fmt.Errorf("invalid pipeline reference: you must provide exactly one key/value combination")
			}
			// anonymous pipelines have a nil key, so we need this special case
			return map[*string]interface{}{
				nil: referenceAsInterfaceMap[nil],
			}, nil
		}
	}
	return value, nil
}

// ParseArguments transfers an unstructured map of arguments into a struct
func ParseArguments(
	middlewareArguments interface{},
	middlewareIdentifier string,
	run *Run,
) bool {
	if middlewareArguments == nil {
		return false
	}
	argumentsCopy := run.ArgumentsCopy()
	argument, ok := argumentsCopy[middlewareIdentifier]
	if !ok || argument == nil {
		return false
	}
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook:  pipelineReferenceDecodeHook,
		ErrorUnused: true,
		Result:      middlewareArguments,
	}
	decoder, _ := mapstructure.NewDecoder(&decoderConfig)
	err := decoder.Decode(argument)
	if err != nil {
		run.Log.Error(fmt.Errorf("malformed arguments for %q: %w", middlewareIdentifier, err))
		return false
	}
	return true
}

// ParseArgumentsIncludingParents is like ParseArguments, but will traverse through parents if no suitable has been found
func ParseArgumentsIncludingParents(
	middlewareArguments interface{},
	middlewareIdentifier string,
	run *Run,
) bool {
	currentRun := run
	for currentRun != nil && !ParseArguments(middlewareArguments, middlewareIdentifier, currentRun) {
		currentRun = currentRun.Parent
	}
	return currentRun != nil
}

func CollectReferences(references []Reference) ([]*string, []map[string]interface{}, []string) {
	childIdentifiers := make([]*string, 0, len(references))
	childArguments := make([]map[string]interface{}, 0, len(references))
	info := make([]string, 0, len(references))
	for _, childReference := range references {
		for pipelineIdentifier, pipelineArguments := range childReference {
			childIdentifiers = append(childIdentifiers, pipelineIdentifier)
			childArguments = append(childArguments, stringmap.CopyMap(pipelineArguments))
			if pipelineIdentifier == nil {
				info = append(info, "anonymous")
			} else {
				info = append(info, *pipelineIdentifier)
			}
		}
	}
	return childIdentifiers, childArguments, info
}
