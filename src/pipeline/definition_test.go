package pipeline

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMergePipelineDefinitions(t *testing.T) {
	definition1 := *NewDefinition(nil, "test1", false, false)
	definition2 := *NewDefinition(nil, "test2", false, false)
	definition3 := *NewDefinition(nil, "test3", false, false)
	definition4 := *NewDefinition(nil, "test4", false, false)
	definition5 := *NewDefinition(nil, "test5", false, false)

	lookup1 := DefinitionsLookup{
		"test1": []Definition{
			definition1,
		},
		"test2": []Definition{
			definition2,
		},
	}
	lookup2 := DefinitionsLookup{
		"test1": []Definition{
			definition3,
			definition4,
		},
		"test3": []Definition{
			definition5,
		},
	}

	result := MergePipelineDefinitions(lookup1, lookup2)
	require.Equal(t, DefinitionsLookup{
		"test1": []Definition{
			definition1,
			definition3,
			definition4,
		},
		"test2": []Definition{
			definition2,
		},
		"test3": []Definition{
			definition5,
		},
	}, result)
}
