package models

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMergePipelineDefinitions(t *testing.T) {
	definition1 := *NewPipelineDefinition(nil, PipelineFile{FileName: "test1"}, false, false)
	definition2 := *NewPipelineDefinition(nil, PipelineFile{FileName: "test2"}, false, false)
	definition3 := *NewPipelineDefinition(nil, PipelineFile{FileName: "test3"}, false, false)
	definition4 := *NewPipelineDefinition(nil, PipelineFile{FileName: "test4"}, false, false)
	definition5 := *NewPipelineDefinition(nil, PipelineFile{FileName: "test5"}, false, false)

	lookup1 := PipelineDefinitionsLookup{
		"test1": []PipelineDefinition{
			definition1,
		},
		"test2": []PipelineDefinition{
			definition2,
		},
	}
	lookup2 := PipelineDefinitionsLookup{
		"test1": []PipelineDefinition{
			definition3,
			definition4,
		},
		"test3": []PipelineDefinition{
			definition5,
		},
	}

	result := MergePipelineDefinitions(lookup1, lookup2)
	require.Equal(t, PipelineDefinitionsLookup{
		"test1": []PipelineDefinition{
			definition1,
			definition3,
			definition4,
		},
		"test2": []PipelineDefinition{
			definition2,
		},
		"test3": []PipelineDefinition{
			definition5,
		},
	}, result)
}
