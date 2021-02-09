package pipeline

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseArgumentsIncludingParents(t *testing.T) {
	parentRun, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{"test"},
	}, nil, nil)
	run, _ := NewRun(nil, map[string]interface{}{}, nil, parentRun)
	reference := make([]interface{}, 0, 1)
	require.Equal(t, true, ParseArgumentsIncludingParents(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}

func TestParseArguments_NilMiddlewareArguments(t *testing.T) {
	require.Equal(t, false, ParseArguments(nil, "test", nil))
}

func TestParseArguments_WithMalformedArguments(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": "test",
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]interface{}{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
}

func TestParseArguments_WithoutArguments(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
}

func TestParseArguments_WithValidArguments(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{"test"},
	}, nil, nil)
	reference := make([]interface{}, 0, 1)
	require.Equal(t, true, ParseArguments(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}

func TestParsePipelineReferences_WithInvalidReference_MapWithNilKey(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[interface{}]interface{}{
				nil: map[string]interface{}{
					"test1": "test1",
				},
				"test2": map[string]interface{}{
					"test2": "test2",
				},
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "invalid pipeline reference")
}

func TestParsePipelineReferences_WithInvalidReference_StringMap(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[string]interface{}{
				"test1": map[string]interface{}{},
				"test2": map[string]interface{}{},
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "invalid pipeline reference")
}

func TestParsePipelineReferences_WithInvalidReference_StringPointerMap(t *testing.T) {
	testKey := "test1"
	otherTestKey := "test2"
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[*string]interface{}{
				&testKey:      map[string]interface{}{},
				&otherTestKey: map[string]interface{}{},
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "invalid pipeline reference")
}

func TestParsePipelineReferences_WithInvalidReference_UnknownType(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			[]interface{}{
				"test",
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
}

func TestParsePipelineReferences_WithMalformedArguments_MapWithNilKey(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[interface{}]interface{}{
				nil: "test",
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
}

func TestParsePipelineReferences_WithMalformedArguments_StringMap(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[string]interface{}{
				"test": "test",
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
}

func TestParsePipelineReferences_WithMalformedArguments_StringPointerMap(t *testing.T) {
	testKey := "test"
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[*string]interface{}{
				&testKey: "test",
			},
		},
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
	require.Equal(t, 1, run.Log.ErrorCount())
	require.Contains(t, run.Log.LastError().Error(), "malformed arguments")
}

func TestParsePipelineReferences_WithNilMiddlewareArguments(t *testing.T) {
	require.Equal(t, false, ParseArguments(nil, "test", nil))
}

func TestParsePipelineReferences_WithNonArrayArguments(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": "test",
	}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
}

func TestParsePipelineReferences_WithoutArguments(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{}, nil, nil)
	require.Equal(t, false, ParseArguments(&[]Reference{}, "test", run))
}

func TestParsePipelineReferences_WithValidReference_MapWithNilKey(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[interface{}]interface{}{
				nil: map[string]interface{}{},
			},
		},
	}, nil, nil)
	reference := make([]Reference, 0, 1)
	require.Equal(t, true, ParseArguments(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}

func TestParsePipelineReferences_WithValidReference_String(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			"another_test",
		},
	}, nil, nil)
	reference := make([]Reference, 0, 1)
	require.Equal(t, true, ParseArguments(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}

func TestParsePipelineReferences_WithValidReference_StringMap(t *testing.T) {
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[string]interface{}{
				"test1": map[string]interface{}{},
			},
		},
	}, nil, nil)
	reference := make([]Reference, 0, 1)
	require.Equal(t, true, ParseArguments(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}

func TestParsePipelineReferences_WithValidReference_StringPointerMap(t *testing.T) {
	test := "test"
	run, _ := NewRun(nil, map[string]interface{}{
		"test": []interface{}{
			map[*string]interface{}{
				&test: map[string]interface{}{},
			},
		},
	}, nil, nil)
	reference := make([]Reference, 0, 1)
	require.Equal(t, true, ParseArguments(&reference, "test", run))
	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, 1, len(reference))
}
