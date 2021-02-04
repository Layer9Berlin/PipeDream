package parsers

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsers_Pipelines_BuiltInPipelineFilePaths_noBuiltInPipelinesFound(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{}, nil
		}),
	)
	result, err := parser.BuiltInPipelineFilePaths("test/invalid/3247285235434")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no built-in pipeline files found")
	require.Equal(t, result, []string{})
}

func TestParsers_Pipelines_BuiltInPipelineFilePaths_globError(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{}, fmt.Errorf("test error")
		}),
	)
	result, err := parser.BuiltInPipelineFilePaths("test/invalid/3247285235434")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to glob pipeline files: test error")
	require.Equal(t, result, []string{})
}

func TestParsers_Pipelines_UserPipelineFilePaths(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths([]string{"test1", "test2.pipe"})
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe", "test2.pipe"}, result)
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths([]string{})
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe", "test2.pipe"}, result)
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs_globError(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, fmt.Errorf("test error")
		}))
	_, err := parser.UserPipelineFilePaths([]string{})
	require.NotNil(t, err)
	require.Equal(t, "failed to glob pipeline files: test error", err.Error())
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs_noMatches(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{}, nil
		}))
	pipelineFilePaths, err := parser.UserPipelineFilePaths([]string{})
	require.Nil(t, err)
	require.Equal(t, []string{}, pipelineFilePaths)
}

func TestParsers_Pipelines_UserPipelineFilePaths_passSinglePathArg(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths([]string{"test1"})
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe"}, result)
}
