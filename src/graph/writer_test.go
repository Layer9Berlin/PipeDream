package graph

import (
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestGraph_OutputGraph(t *testing.T) {
	graphWriter := NewWriter()
	graphWriter.OpenInBrowser = func(file string) error {
		return nil
	}
	executionContext := middleware.NewExecutionContext()
	err := graphWriter.Write(executionContext)

	require.Nil(t, err)
}

func TestGraph_OutputGraph_templateError(t *testing.T) {
	graphWriter := NewWriter()
	graphWriter.NewTemplate = func() (*template.Template, error) {
		return nil, fmt.Errorf("test error")
	}
	executionContext := middleware.NewExecutionContext()
	err := graphWriter.Write(executionContext)

	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
}

func TestGraph_OutputGraph_tempfileError(t *testing.T) {
	graphWriter := NewWriter()
	graphWriter.TempFile = func(dir string, pattern string) (*os.File, error) {
		return nil, fmt.Errorf("test error")
	}
	executionContext := middleware.NewExecutionContext()
	err := graphWriter.Write(executionContext)

	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
}

func TestGraph_OutputGraph_executeError(t *testing.T) {
	graphWriter := NewWriter()
	graphWriter.Execute = func(template *template.Template, wr io.Writer, data interface{}) error {
		return fmt.Errorf("test error")
	}
	executionContext := middleware.NewExecutionContext()
	err := graphWriter.Write(executionContext)

	require.NotNil(t, err)
	require.Equal(t, "test error", err.Error())
}

func TestGraph_OutputGraph_tempfileCloseError(t *testing.T) {
	graphWriter := NewWriter()
	graphWriter.TempFile = func(dir string, pattern string) (*os.File, error) {
		// create a tempfile and close it, so that the next close will error out
		file, _ := ioutil.TempFile(os.TempDir(), "*.html")
		_ = file.Close()
		return file, nil
	}
	executionContext := middleware.NewExecutionContext()
	err := graphWriter.Write(executionContext)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "file already closed")
}
