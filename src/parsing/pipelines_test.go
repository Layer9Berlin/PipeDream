package parsing

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
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
	result, err := parser.UserPipelineFilePaths("")
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe", "test2.pipe"}, result)
}

func TestParsers_Pipelines_UserPipelineFilePaths_overrider(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths("test/test3.pipe")
	require.Nil(t, err)
	require.Equal(t, []string{"test/test3.pipe"}, result)
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths("")
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe", "test2.pipe"}, result)
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs_globError(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, fmt.Errorf("test error")
		}))
	_, err := parser.UserPipelineFilePaths("")
	require.NotNil(t, err)
	require.Equal(t, "failed to glob pipeline files: test error", err.Error())
}

func TestParsers_Pipelines_UserPipelineFilePaths_noArgs_noMatches(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{}, nil
		}))
	pipelineFilePaths, err := parser.UserPipelineFilePaths("")
	require.Nil(t, err)
	require.Equal(t, []string{}, pipelineFilePaths)
}

func TestParsers_Pipelines_UserPipelineFilePaths_passSinglePathArg(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe"}, nil
		}))
	result, err := parser.UserPipelineFilePaths("test1")
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe"}, result)
}

func TestParsers_Pipelines_RecursivelyAddImportsImplementation(t *testing.T) {
	parser := NewParser(
		WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
			return []string{"test1.pipe"}, nil
		}))
	result, err := parser.RecursivelyAddImports([]string{"test2"})
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe"}, result)
}

func TestPipelines_BuiltInPipelineFilePaths_nonPathError(t *testing.T) {
	parser := NewParser(
		WithEvalSymlinksImplementation(func(path string) (string, error) {
			return "", fmt.Errorf("test error")
		}),
	)
	_, err := parser.BuiltInPipelineFilePaths("test")
	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
}

func TestPipelines_BuiltInPipelineFilePaths_nonENOENTPathError(t *testing.T) {
	parser := NewParser(
		WithEvalSymlinksImplementation(func(path string) (string, error) {
			return "", &os.PathError{Err: fmt.Errorf("test error")}
		}),
	)
	_, err := parser.BuiltInPipelineFilePaths("test")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
}

func TestPipelines_BuiltInPipelineFilePaths(t *testing.T) {
	parser := NewParser(
		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
			return []string{"test1.pipe", "test2.pipe"}, nil
		}),
		WithEvalSymlinksImplementation(func(path string) (string, error) {
			return "", nil
		}),
	)
	matches, err := parser.BuiltInPipelineFilePaths("test")
	require.Nil(t, err)
	require.Equal(t, []string{"test1.pipe", "test2.pipe"}, matches)
}
