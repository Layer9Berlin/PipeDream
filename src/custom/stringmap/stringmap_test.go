package stringmap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCopyMap_WithSuccess(t *testing.T) {
	map1 := StringMap{
		"test1": "test3",
		"test2": map[string]interface{}{
			"test4": StringMap{
				"test5": "test6",
			},
		},
		"test7": []interface{}{
			interface{}("test8"),
			interface{}([]string{
				"test10",
				"test11",
			}),
			interface{}(StringMap{
				"test12": nil,
			}),
		},
		"test13": map[interface{}]interface{}{
			"test": "value",
		},
	}

	map2 := CopyMap(map1)

	map1["test1"] = "new"
	map1["test2"].(StringMap)["test4"].(StringMap)["test5"] = "new2"
	delete(map1["test7"].([]interface{})[2].(StringMap), "test12")

	require.Equal(t, "new", map1["test1"])
	require.Equal(t, StringMap{
		"test4": StringMap{
			"test5": "new2",
		},
	}, map1["test2"])
	require.Equal(t, StringMap{}, map1["test7"].([]interface{})[2])

	require.Equal(t, "test3", map2["test1"])
	require.Equal(t, StringMap{
		"test4": StringMap{
			"test5": "test6",
		},
	}, map2["test2"])
	require.Equal(t, StringMap{
		"test12": nil,
	}, map2["test7"].([]interface{})[2])
	require.Equal(t, map[interface{}]interface{}{
		"test": "value",
	}, map2["test13"])
}

func TestCopyMap_Nil(t *testing.T) {
	copiedMap := CopyMap(nil)

	require.Equal(t, map[string]interface{}{}, copiedMap)
}

func TestMergeIntoMap_WithSuccess(t *testing.T) {
	map1 := StringMap{
		"test1": "test3",
		"test2": map[string]interface{}{
			"test4": StringMap{
				"test5": "test6",
			},
		},
		"test7": []interface{}{
			interface{}("test8"),
			interface{}([]string{
				"test10",
				"test11",
			}),
			interface{}(StringMap{
				"test12": nil,
			}),
		},
	}

	map2 := StringMap{
		"test1": "new",
		"test2": map[string]interface{}{
			"test4": StringMap{
				"test5":  "new2",
				"test22": "test23",
			},
		},
		"test7": []interface{}{
			interface{}("test9"),
			interface{}(StringMap{
				"test13": "test14",
			}),
		},
		"test15": "test16",
		"test17": StringMap{
			"test18": "test19",
		},
		"test20": []interface{}{
			interface{}("test21"),
		},
	}

	err := MergeIntoMap(map1, map2)

	require.Nil(t, err)
	require.Equal(t, StringMap{
		"test1": "test3",
		"test2": map[string]interface{}{
			"test4": StringMap{
				"test5":  "test6",
				"test22": "test23",
			},
		},
		"test7": []interface{}{
			interface{}("test8"),
			interface{}([]string{
				"test10",
				"test11",
			}),
			interface{}(StringMap{
				"test12": nil,
			}),
		},
		"test15": "test16",
		"test17": StringMap{
			"test18": "test19",
		},
		"test20": []interface{}{
			interface{}("test21"),
		},
	}, map1)

	// test we are actually deling with a copy, not the original map
	delete(map1["test2"].(StringMap)["test4"].(StringMap), "test22")
	require.Equal(t, map2["test2"].(StringMap)["test4"].(StringMap)["test22"], "test23")
	delete(map1["test17"].(StringMap), "test18")
	require.Equal(t, map2["test17"].(StringMap)["test18"], "test19")
}

func TestMergeIntoMap_WithErrors(t *testing.T) {
	// one of the maps being nil should be handled gracefully
	// after all, this will often be the case for pipe arguments
	require.Nil(t, MergeIntoMap(nil, StringMap{
		"test1": "test1",
	}))
	require.Nil(t, MergeIntoMap(StringMap{
		"test1": "test1",
	}, nil))

	// incompatible maps should produce an error
	require.NotNil(t, MergeIntoMap(StringMap{
		"test1": "test2",
	}, StringMap{
		"test1": StringMap{
			"test3": "test4",
		},
	}))
	require.NotNil(t, MergeIntoMap(StringMap{
		"test1": StringMap{
			"test3": "test4",
		},
	}, StringMap{
		"test1": "test2",
	}))
}

func TestGetValueInMap_Nested(t *testing.T) {
	vector := map[string]interface{}{
		"test1": map[string]interface{}{
			"test2": map[string]interface{}{
				"test3": map[string]interface{}{
					"test4": "test5",
				},
			},
		},
	}
	value, err := GetValueInMap(vector, "test1", "test2", "test3", "test4")
	require.Nil(t, err)
	require.Equal(t, "test5", value)
}

func TestGetValueInMap_NotExisting(t *testing.T) {
	vector := map[string]interface{}{}
	_, err := GetValueInMap(vector, "impossible")
	require.NotNil(t, err)
	require.Equal(t, "value does not exist at path", err.Error())
}

func TestGetValueInMap_InvalidMap(t *testing.T) {
	vector := map[string]interface{}{
		"test": []string{"invalid"},
	}
	_, err := GetValueInMap(vector, "test", "impossible")
	require.NotNil(t, err)
	require.Equal(t, "value does not exist at path", err.Error())
}

func TestSetValueInMap_InvalidMap(t *testing.T) {
	vector := map[string]interface{}{
		"test": []string{"invalid"},
	}
	err := SetValueInMap(vector, "new", "test", "impossible")
	require.NotNil(t, err)
	require.Equal(t, "failed to set new value, encountered something other than a string map", err.Error())
}

func TestSetValueInMap_NonExistentLeaf(t *testing.T) {
	vector := map[string]interface{}{
		"test1": map[string]interface{}{
			"test2": map[string]interface{}{},
		},
	}
	err := SetValueInMap(vector, "test5", "test1", "test2", "test3", "test4")
	require.Nil(t, err)
	value, err := GetValueInMap(vector, "test1", "test2", "test3", "test4")
	require.Nil(t, err)
	require.Equal(t, "test5", value)
}
