package datastream

import (
	"io"
)

// ComposableDataStreamIntercept is a structure for intercepting and manipulating data passed through a data stream
type ComposableDataStreamIntercept struct {
	reader       io.Reader
	writer       io.WriteCloser
	errorHandler func(error)
}

func newComposableDataStreamIntercept(reader io.Reader, writer io.WriteCloser, errorHandler func(error)) *ComposableDataStreamIntercept {
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

// Close closes the data stream intercept to indicate that all data has been written
func (intercept *ComposableDataStreamIntercept) Close() error {
	return intercept.writer.Close()
}
