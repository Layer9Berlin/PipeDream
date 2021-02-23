package graph

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Writer struct {
	OpenInBrowser func(file string) error
	NewTemplate   func() (*template.Template, error)
	TempFile      func(dir, pattern string) (f *os.File, err error)
	Execute       func(template *template.Template, wr io.Writer, data interface{}) error
}

func NewWriter() *Writer {
	return &Writer{
		NewTemplate: func() (*template.Template, error) {
			return template.New("runs").Parse(graphTemplate())
		},
		TempFile: ioutil.TempFile,
		Execute: func(template *template.Template, wr io.Writer, data interface{}) error {
			return template.Execute(wr, data)
		},
		OpenInBrowser: func(file string) error {
			return exec.Command("open", file).Run()
		},
	}
}

func (writer *Writer) Write(executionContext interface{}) error {
	ut, err := writer.NewTemplate()
	if err != nil {
		return err
	}

	tmpFile, err := writer.TempFile(os.TempDir(), "*.html")
	if err != nil {
		return err
	}

	err = writer.Execute(ut, tmpFile, executionContext)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	return writer.OpenInBrowser(tmpFile.Name())
}
