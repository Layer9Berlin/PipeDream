package middleware

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/models"
	"github.com/Layer9Berlin/pipedream/src/parsers"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestExecutionContext_CancelAll(t *testing.T) {
	executionContext := NewExecutionContext()
	executionContext.Runs = []*models.PipelineRun{{}, {}}
	err := executionContext.CancelAll()
	require.Nil(t, err)
	require.True(t, executionContext.Runs[0].Cancelled())
	require.True(t, executionContext.Runs[1].Cancelled())
}

func TestExecutionContext_FullRun_WithoutOptions(t *testing.T) {
	executionContext := NewExecutionContext()
	run := executionContext.FullRun()
	require.NotNil(t, run)
	require.Equal(t, run, executionContext.rootRun)
}

func TestExecutionContext_FullRun_WithDefinitionsLookupOption(t *testing.T) {
	arguments := map[string]interface{}{
		"key": "value",
	}
	executionContext := NewExecutionContext(WithDefinitionsLookup(map[string][]models.PipelineDefinition{
		"test": {
			{
				DefinitionArguments: arguments,
			},
		},
	}))
	identifier := "test"
	run := executionContext.FullRun(WithIdentifier(&identifier))
	require.NotNil(t, run)
	require.Equal(t, arguments, run.ArgumentsCopy())
}

func TestExecutionContext_FullRun_WithUnmergeableArguments(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("failed to encounter expected panic")
		}
	}()
	arguments1 := map[string]interface{}{
		"key": []interface{}{},
	}
	arguments2 := map[string]interface{}{
		"key": map[string]interface{}{},
	}
	executionContext := ExecutionContext{
		Definitions: map[string][]models.PipelineDefinition{
			"test": {
				{
					DefinitionArguments: arguments1,
				},
			},
		},
	}
	identifier := "test"
	run := executionContext.FullRun(WithIdentifier(&identifier), WithArguments(arguments2))
	require.Nil(t, run)
}

func TestExecutionContext_FullRun_WithActivityIndicator(t *testing.T) {
	executionContext := NewExecutionContext(WithActivityIndicator(logging.NewNestedActivityIndicator()))
	identifier := "test"
	run := executionContext.FullRun(WithIdentifier(&identifier))
	require.NotNil(t, run)
	require.Equal(t, 1, executionContext.ActivityIndicator.Len())
}

func TestExecutionContext_FullRun_WithSetupFunction(t *testing.T) {
	setupCalled := false
	setupFunc := func(run *models.PipelineRun) {
		setupCalled = true
	}
	executionContext := NewExecutionContext()
	identifier := "test"
	run := executionContext.FullRun(
		WithIdentifier(&identifier),
		WithSetupFunc(setupFunc),
	)
	require.NotNil(t, run)
	require.True(t, setupCalled)
}

func TestExecutionContext_UnwindStack(t *testing.T) {
	middleware1 := NewFakeMiddleware()
	middleware2 := NewFakeMiddleware()
	stack := []Middleware{
		middleware1,
		middleware2,
	}

	executionContext := ExecutionContext{
		MiddlewareStack: stack,
	}
	executionContext.unwindStack(nil, 0)
	require.Equal(t, 1, middleware1.CallCount)
	require.Equal(t, 1, middleware2.CallCount)
}

func TestExecutionContext_PipelineFileAtPath(t *testing.T) {
	executionContext := ExecutionContext{
		PipelineFiles: []models.PipelineFile{
			{
				FileName: "test1",
			},
			{
				FileName: "test2",
			},
			{
				FileName: "test3",
			},
		},
	}
	file, err := executionContext.PipelineFileAtPath("test2")
	require.Nil(t, err)
	require.NotNil(t, file)
}

func TestExecutionContext_PipelineFileAtPath_NotFound(t *testing.T) {
	executionContext := ExecutionContext{
		PipelineFiles: []models.PipelineFile{
			{
				FileName: "test1",
			},
			{
				FileName: "test2",
			},
			{
				FileName: "test3",
			},
		},
	}
	file, err := executionContext.PipelineFileAtPath("test4")
	require.NotNil(t, err)
	require.Nil(t, file)
}

func TestExecutionContext_LookUpPipelineDefinition(t *testing.T) {
	definitionsLookup := map[string][]models.PipelineDefinition{
		"test1": {
			{
				FileName: "test1.file",
				Public:   false,
			},
			{
				FileName: "test2.file",
				Public:   true,
			},
			{
				FileName: "test3.file",
				Public:   false,
			},
		},
		"test2": {},
	}
	definition, found := LookUpPipelineDefinition(definitionsLookup, "test1", "test3.file")
	require.True(t, found)
	require.Equal(t, "test3.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test1", "test4.file")
	require.True(t, found)
	require.Equal(t, "test2.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test1", "test1.file")
	require.True(t, found)
	require.Equal(t, "test1.file", definition.FileName)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "test2", "test1.file")
	require.False(t, found)
	require.Nil(t, definition)

	definition, found = LookUpPipelineDefinition(definitionsLookup, "invalid", "test1.file")
	require.False(t, found)
	require.Nil(t, definition)
}

func TestExecutionContext_Execute(t *testing.T) {
	buffer := new(bytes.Buffer)
	executionContext := NewExecutionContext()
	executionContext.Execute("test", buffer)
	require.NotContains(t, buffer.String(), "====== LOGS ======")
	require.Contains(t, buffer.String(), "===== RESULT =====")
	require.NotContains(t, buffer.String(), "===== ERRORS =====")
}

func TestExecutionContext_SetUpPipelines(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(
					func(pattern string) ([]string, error) {
						return []string{
							"test1.pipe",
							"test2.pipe",
						}, nil
					}),
				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte(""), nil
				}))))
	buffer := new(bytes.Buffer)
	executionContext.Log.SetOutput(buffer)

	err := executionContext.SetUpPipelines([]string{})
	require.Nil(t, err)
	require.NotNil(t, executionContext.PipelineFiles)

	require.Equal(t, "", buffer.String())
}

func TestExecutionContext_SetUpPipelines_BuiltInPipelineFilePathsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{}, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines([]string{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_ParseBuiltInPipelineFilesError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return nil, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines([]string{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_UserPipelineFilePathsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					if strings.Contains(pattern, "pipes/**") {
						return []string{"test.file"}, nil
					}
					return []string{}, fmt.Errorf("test error")
				}),
				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte{}, nil
				}),
			)))
	err := executionContext.SetUpPipelines([]string{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_RecursivelyAddImportsError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
					return []byte{}, nil
				}),
				parsers.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return nil, fmt.Errorf("test error")
				}),
			)))
	err := executionContext.SetUpPipelines([]string{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

func TestExecutionContext_SetUpPipelines_ParsePipelineFilesError(t *testing.T) {
	executionContext := NewExecutionContext(
		WithParser(
			parsers.NewParser(
				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
					return []string{"test1"}, nil
				}),
				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
					if filename == "test.file" {
						return nil, fmt.Errorf("test error")
					}
					return []byte{}, nil
				}),
				parsers.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
					return []string{"test.file"}, nil
				}),
			)))
	err := executionContext.SetUpPipelines([]string{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test error")
	require.Nil(t, executionContext.PipelineFiles)
}

//
//func TestExecutionContext_SetUpPipelines_ParsePipelineFilesError(t *testing.T) {
//	executionContext := NewExecutionContext(
//		WithParser(
//			parsers.NewParser(
//				parsers.WithFindByGlobImplementation(func(pattern string) ([]string, error) {
//					if pattern == "pipes/**/*.pipe" {
//						return []string{"test2.pipe"}, nil
//					}
//					return []string{"test1.pipe"}, nil
//				}),
//				parsers.WithReadFileImplementation(func(filename string) ([]byte, error) {
//					if filename == "test.file" {
//						return nil, fmt.Errorf("test error")
//					}
//					return []byte{}, nil
//				}),
//				parsers.WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
//					return []string{"test.file"}, nil
//				}))),
//		WithDefinitionsLookup(models.PipelineDefinitionsLookup{}),
//	)
//	err := executionContext.SetUpPipelines([]string{"test1"})
//	require.NotNil(t, err)
//	require.Contains(t, err.Error(), "test error")
//	require.Equal(t, []models.PipelineFile{}, executionContext.PipelineFiles)
//}

//func TestRun_PipelineSetupHelper_setUpPipelines(t *testing.T) {
//	executionContext := middleware.NewExecutionContext(
//		middleware.WithDefinitionsLookup(models.PipelineDefinitionsLookup{}),
//		WithFindByGlobImplementation(func(pattern string) ([]string, error) {
//			return []string{"test1.pipe"}, nil
//		}),
//		WithRecursivelyAddImportsImplementation(func(paths []string) ([]string, error) {
//			return []string{"test2.pipe", "test3.pipe"}, nil
//		}),
//		WithParsePipelinesImplementation(func(allPipelineFilePaths []string, builtIn bool, executionContext *middleware.ExecutionContext) error {
//			return nil
//		}),
//	)
//	err := executionContext.SetUpPipelines([]string{"test1"})
//	require.Nil(t, err)
//	require.Equal(t, []models.PipelineFile{}, executionContext.PipelineFiles)
//}

type FakeMiddleware struct {
	CallCount int
}

func NewFakeMiddleware() *FakeMiddleware {
	return &FakeMiddleware{
		CallCount: 0,
	}
}

func (fakeMiddleware *FakeMiddleware) String() string {
	return "fake"
}

func (fakeMiddleware *FakeMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *ExecutionContext,
) {
	fakeMiddleware.CallCount += 1
	next(run)
}
