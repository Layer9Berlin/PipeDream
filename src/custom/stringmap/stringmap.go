// Package stringmap contains convenience functions for maps with string keys
package stringmap

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
)

// StringMap is a map with string keys and arbitrary values
type StringMap = map[string]interface{}

// CopyMap creates a deep copy of a StringMap
//
// The copy can be safely mutated without changing nested keys of the original StringMap.
func CopyMap(otherMap StringMap) StringMap {
	if otherMap == nil {
		return make(StringMap, 0)
	}
	result := make(StringMap, len(otherMap))
	for otherMapKey, otherMapValue := range otherMap {
		result[otherMapKey] = copyValue(otherMapValue)
	}
	return result
}

func copyValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	switch typedValue := value.(type) {
	case StringMap:
		result := StringMap{}
		for stringKey, subValue := range typedValue {
			result[stringKey] = copyValue(subValue)
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(typedValue))
		for _, subValue := range typedValue {
			result = append(result, copyValue(subValue))
		}
		return result
	case map[interface{}]interface{}:
		result := map[interface{}]interface{}{}
		for interfaceKey, subValue := range typedValue {
			result[interfaceKey] = copyValue(subValue)
		}
		return result
	default:
		return value
	}
}

// MergeIntoMap deep merges values from one StringMap into another
//
// For each key path at which subject does not have a value,
// this will set the value from otherMap
// even if subject has a value at a prefix of the key path.
// Existing values - including nil - at matching complete paths will not be overwritten.
func MergeIntoMap(subject StringMap, otherMap StringMap) error {
	if subject == nil || otherMap == nil {
		return nil
	}
	var allErrors *multierror.Error
	for otherMapKey, otherMapValue := range otherMap {
		existingValue, haveExistingValue := subject[otherMapKey]
		if haveExistingValue {
			existingStringMapValue, existingValueIsStringMap := existingValue.(StringMap)
			otherMapValueAsStringMap, otherMapValueIsStringMap := otherMapValue.(StringMap)
			if existingValueIsStringMap {
				if otherMapValueIsStringMap {
					allErrors = multierror.Append(allErrors, MergeIntoMap(existingStringMapValue, otherMapValueAsStringMap))
				} else {
					allErrors = multierror.Append(allErrors, fmt.Errorf("cannot merge value of type %T into existing value of type %T", otherMapValue, existingValue))
				}
			} else {
				if otherMapValueIsStringMap {
					allErrors = multierror.Append(allErrors, fmt.Errorf("cannot merge value of type %T into existing value of type %T", otherMapValue, existingValue))
				}
			}
			// nothing to do if the existing value is not a StringMap,
			// we won't replace it, even if it is nil
			// this allows removing arguments set on an outer pipe
			// by setting the value to `null` on the inner pipe
		} else {
			subject[otherMapKey] = copyValue(otherMapValue)
		}
	}
	if allErrors == nil {
		return nil
	}
	return allErrors.ErrorOrNil()
}

// GetValueInMap returns the value of a nested StringMap at the specified path
func GetValueInMap(searchedMap map[string]interface{}, path ...string) (interface{}, error) {
	firstComponent, restOfPath := path[0], path[1:]
	existingValue, haveExistingValue := searchedMap[firstComponent]
	if haveExistingValue {
		if len(restOfPath) == 0 {
			return existingValue, nil
		}

		nestedMap, haveNestedMap := existingValue.(map[string]interface{})
		if haveNestedMap {
			return GetValueInMap(nestedMap, restOfPath...)
		}
		return nil, fmt.Errorf("value does not exist at path")
	}
	return nil, fmt.Errorf("value does not exist at path")
}

// HaveValueInMap indicates whether the nested StringMap has a value at the specified path
func HaveValueInMap(searchedMap map[string]interface{}, path ...string) bool {
	if len(path) == 0 {
		return false
	}
	firstComponent, restOfPath := path[0], path[1:]
	existingValue, haveExistingValue := searchedMap[firstComponent]
	if haveExistingValue {
		if len(restOfPath) == 0 {
			return true
		}

		nestedMap, haveNestedMap := existingValue.(map[string]interface{})
		if haveNestedMap {
			return HaveValueInMap(nestedMap, restOfPath...)
		}
		return false
	}
	return false
}

// SetValueInMap fixes the value of a nested StringMap at the specified path
//
// Additional levels of nesting will be created if necessary.
func SetValueInMap(mapToChange map[string]interface{}, value interface{}, path ...string) error {
	firstComponent, restOfPath := path[0], path[1:]
	nextValue, haveNextValue := mapToChange[firstComponent]
	if !haveNextValue {
		nextValue = make(map[string]interface{}, 1)
		mapToChange[firstComponent] = nextValue
	}
	if len(restOfPath) == 0 {
		mapToChange[firstComponent] = value
		return nil
	}

	nestedMap, haveNestedMap := nextValue.(map[string]interface{})
	if haveNestedMap {
		return SetValueInMap(nestedMap, value, restOfPath...)
	}
	return fmt.Errorf("failed to set new value, encountered something other than a string map")
}

// RemoveValueInMap removes the value of a nested StringMap at the specified path
//
// Trying to remove a non-existent value returns an error.
func RemoveValueInMap(mapToChange map[string]interface{}, path ...string) error {
	firstComponent, restOfPath := path[0], path[1:]
	nextValue, haveNextValue := mapToChange[firstComponent]
	if !haveNextValue {
		return fmt.Errorf("failed to remove value, as it could not be found")
	}
	if len(restOfPath) == 0 {
		delete(mapToChange, firstComponent)
		return nil
	}

	nestedMap, haveNestedMap := nextValue.(map[string]interface{})
	if haveNestedMap {
		return RemoveValueInMap(nestedMap, restOfPath...)
	}
	return fmt.Errorf("failed to remove value, encountered something other than a string map")
}
