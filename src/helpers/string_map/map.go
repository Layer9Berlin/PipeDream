package string_map

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
)

type StringMap = map[string]interface{}

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
		for stringKey, subValue := range typedValue {
			result[stringKey] = copyValue(subValue)
		}
		return result
	default:
		return value
	}
}

// merge values from another map into the subject:
// for each key path at which the subject does not have a value,
// this will set the value from the other map
// even if the subject has a value at a prefix of the key path
// existing values - including nil - will not be overwritten
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

func GetValueInMap(searchedMap map[string]interface{}, path ...string) (interface{}, error) {
	firstComponent, restOfPath := path[0], path[1:]
	existingValue, haveExistingValue := searchedMap[firstComponent]
	if haveExistingValue {
		if len(restOfPath) == 0 {
			return existingValue, nil
		} else {
			nestedMap, haveNestedMap := existingValue.(map[string]interface{})
			if haveNestedMap {
				return GetValueInMap(nestedMap, restOfPath...)
			} else {
				return nil, fmt.Errorf("value does not exist at path")
			}
		}
	} else {
		return nil, fmt.Errorf("value does not exist at path")
	}
}

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
	} else {
		nestedMap, haveNestedMap := nextValue.(map[string]interface{})
		if haveNestedMap {
			return SetValueInMap(nestedMap, value, restOfPath...)
		} else {
			return fmt.Errorf("failed to set new value, encountered something other than a string map")
		}
	}
}
