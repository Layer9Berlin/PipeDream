package datastream

import (
	"io"
)

type ComposableDataStreamIntercept struct {
	reader       io.Reader
	writer       io.WriteCloser
	errorHandler func(error)
}

func NewComposableDataStreamIntercept(reader io.Reader, writer io.WriteCloser, errorHandler func(error)) *ComposableDataStreamIntercept {
	return &ComposableDataStreamIntercept{
		reader:       reader,
		writer:       writer,
		errorHandler: errorHandler,
	}
}

func (intercept *ComposableDataStreamIntercept) Read(p []byte) (int, error) {
	return intercept.reader.Read(p)
}

func (intercept *ComposableDataStreamIntercept) Write(p []byte) (int, error) {
	return intercept.writer.Write(p)
}

func (intercept *ComposableDataStreamIntercept) Close() error {
	return intercept.writer.Close()
}
