package parsers

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_ParsePipelineFiles(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			switch filename {
			case "file1":
				return []byte(`
version: 0.0.1

default:
  command: test-command
  dir: test-dir

public:
  test:
    description: Public test pipe
    pipe:
      - test1:
          arg1: value1
      - test2

private:
  test1:
    pipe:
      - run:
          command: "test1"

  test2:
    pipe:
      - run:
          command: "test2"
`), nil
			case "file2":
				return []byte(`
version: 0.0.2

public:
  test:
    key1: value1

private:
  test:
    key2: value2
`), nil
			default:
				return nil, nil
			}
		}))
	defaults, definitions, files, err := parser.ParsePipelineFiles([]string{"file1", "file2"}, false)
	require.Nil(t, err)
	require.Equal(t, models.DefaultSettings{Command: "test-command", Dir: "test-dir"}, defaults)
	require.Equal(t, map[string][]models.PipelineDefinition{
		"test": {
			{
				BuiltIn: false,
				DefinitionArguments: map[string]interface{}{
					"description": "Public test pipe",
					"pipe": []interface{}{
						map[string]interface{}{
							"test1": map[string]interface{}{
								"arg1": "value1",
							},
						},
						"test2",
					},
				},
				FileName: "file1",
				Public:   true,
			},
			{
				BuiltIn: false,
				DefinitionArguments: map[string]interface{}{
					"key1": "value1",
				},
				FileName: "file2",
				Public:   true,
			},
			{
				BuiltIn: false,
				DefinitionArguments: map[string]interface{}{
					"key2": "value2",
				},
				FileName: "file2",
				Public:   false,
			},
		},
		"test1": {
			{
				BuiltIn: false,
				DefinitionArguments: map[string]interface{}{
					"pipe": []interface{}{
						map[string]interface{}{
							"run": map[string]interface{}{
								"command": "test1",
							},
						},
					},
				},
				FileName: "file1",
				Public:   false,
			},
		},
		"test2": {
			{
				BuiltIn: false,
				DefinitionArguments: map[string]interface{}{
					"pipe": []interface{}{
						map[string]interface{}{
							"run": map[string]interface{}{
								"command": "test2",
							},
						},
					},
				},
				FileName: "file1",
				Public:   false,
			},
		},
	}, definitions)
	require.Equal(t, []models.PipelineFile{
		{
			Default:  models.DefaultSettings{Command: "test-command", Dir: "test-dir"},
			FileName: "file1",
			Public: map[string]map[string]interface{}{
				"test": {
					"description": "Public test pipe",
					"pipe": []interface{}{
						map[string]interface{}{
							"test1": map[string]interface{}{
								"arg1": "value1",
							},
						}, "test2"},
				},
			},
			Private: map[string]map[string]interface{}{
				"test1": {
					"pipe": []interface{}{
						map[string]interface{}{
							"run": map[string]interface{}{
								"command": "test1",
							},
						},
					},
				},
				"test2": {
					"pipe": []interface{}{
						map[string]interface{}{
							"run": map[string]interface{}{
								"command": "test2",
							},
						},
					},
				},
			},
		},
		{
			Default:  models.DefaultSettings{Command: "", Dir: ""},
			FileName: "file2",
			Public: map[string]map[string]interface{}{
				"test": {
					"key1": "value1",
				},
			},
			Private: map[string]map[string]interface{}{
				"test": {
					"key2": "value2",
				},
			},
		},
	}, files)
}

func TestReadFileError(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			return []byte{}, fmt.Errorf("test error")
		}))

	_, _, _, err := parser.ParsePipelineFiles([]string{"file"}, false)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
}

func TestParseFileError(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			return []byte("version: 0.0.1\n\n\ntest  - unparseable\n\n"), nil
		}))

	_, _, _, err := parser.ParsePipelineFiles([]string{"file"}, false)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "line 4")
}

func TestInvalidPipelineDefinition(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			return []byte(`
version: 0.0.1

default:
 command: test-command
 dir: test-dir

public:
 test:
   - this is not valid
	- at all
`), nil
		}))

	_, _, _, err := parser.ParsePipelineFiles([]string{"file"}, false)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "line 10")
}

func TestParser_ProcessPipelineFile(t *testing.T) {
	parser := NewParser()
	pipelineFile := models.PipelineFile{
		Public: map[string]map[string]interface{}{
			"test1": {
				"name": "Test 1",
				"pipe": []interface{}{"test2", "test3", map[string]interface{}{
					"test4": interface{}(map[string]interface{}{
						"option1": "value1",
						"option2": "value2",
					}),
				},
				},
			},
		},
	}

	definitions := parser.ProcessPipelineFile(pipelineFile, false)
	require.Equal(t, models.PipelineDefinitionsLookup{
		"test1": []models.PipelineDefinition{
			{
				DefinitionArguments: map[string]interface{}{
					"name": "Test 1",
					"pipe": []interface{}{"test2", "test3", map[string]interface{}{
						"test4": interface{}(map[string]interface{}{
							"option1": "value1",
							"option2": "value2",
						}),
					},
					},
				},
				Public: true,
			},
		},
	}, definitions)
}

func TestParser_Imports(t *testing.T) {
	importedFiles := make([]string, 0, 2)
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			importedFiles = append(importedFiles, filename)
			switch filename {
			case "test1.file":
				return []byte(`
version: 0.0.1

import:
 - test2.file
`), nil
			case "test2.file":
				return []byte(`
version: 0.0.1
`), nil
			}
			require.Fail(t, "unexpected import")
			return nil, nil
		}))

	imports, err := parser.RecursivelyAddImports([]string{
		"test1.file",
		"test2.file",
	})
	require.Nil(t, err)
	require.Equal(t, []string{
		"test1.file",
		"test2.file",
	}, imports)
	require.Equal(t, []string{
		"test1.file",
		"test2.file",
	}, importedFiles)
}

func TestParser_ProcessPipelineFile_WithMultipleDefinitionsInSameFile(t *testing.T) {
	parser := NewParser()
	pipelineFile := models.PipelineFile{
		Public: map[string]map[string]interface{}{
			"test": {
				"name": "Test 1",
			},
		},
		Private: map[string]map[string]interface{}{
			"test": {
				"name": "Test 2",
			},
		},
	}
	definitions := parser.ProcessPipelineFile(pipelineFile, false)
	require.Equal(t, models.PipelineDefinitionsLookup{
		"test": []models.PipelineDefinition{
			{
				DefinitionArguments: map[string]interface{}{
					"name": "Test 1",
				},
				Public: true,
			},
			{
				DefinitionArguments: map[string]interface{}{
					"name": "Test 2",
				},
				Public: false,
			},
		},
	}, definitions)
}
