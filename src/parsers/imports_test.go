package parsers

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImports_AlreadyProcessedIsIgnored(t *testing.T) {
	importedFiles := []string{}
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			importedFiles = append(importedFiles, filename)
			switch filename {
			case "test1.file":
				return []byte(`
version: 0.0.1

import:
 - test2.file
 - test3.file
`), nil
			case "test2.file":
				return []byte(`
version: 0.0.1
`), nil
			case "test3.file":
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
		"test3.file",
	}, imports)
	require.Equal(t, []string{
		"test1.file",
		"test2.file",
		"test3.file",
	}, importedFiles)
}

func TestImports_FileReadError(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			return nil, fmt.Errorf("file read error")
		}))
	imports, err := parser.RecursivelyAddImports([]string{
		"test1.file",
		"test2.file",
	})
	require.NotNil(t, err)
	require.Equal(t, "file read error", err.Error())
	require.Nil(t, imports)
}

func TestImports_InvalidYaml(t *testing.T) {
	parser := NewParser(
		WithReadFileImplementation(func(filename string) ([]byte, error) {
			return []byte(`
 - invalid - yaml
`), nil
		}))
	imports, err := parser.RecursivelyAddImports([]string{
		"test1.file",
		"test2.file",
	})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "error parsing")
	require.Nil(t, imports)
}
