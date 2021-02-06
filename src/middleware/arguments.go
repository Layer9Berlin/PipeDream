package middleware

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type Arguments = map[string]interface{}
type PipelineReference = map[*string]Arguments

var pipelineReferenceType = reflect.TypeOf(PipelineReference{})

func pipelineReferenceDecodeHook(
	_ reflect.Type,
	outputType reflect.Type,
	value interface{},
) (interface{}, error) {
	if outputType == pipelineReferenceType {
		if valueAsString, valueIsString := value.(string); valueIsString {
			return PipelineReference{
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

func ParseArguments(
	middlewareArguments interface{},
	middlewareIdentifier string,
	run *pipeline.Run,
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

func ParseArgumentsIncludingParents(
	middlewareArguments interface{},
	middlewareIdentifier string,
	run *pipeline.Run,
) bool {
	currentRun := run
	for currentRun != nil && !ParseArguments(middlewareArguments, middlewareIdentifier, currentRun) {
		currentRun = currentRun.Parent
	}
	return currentRun != nil
}
