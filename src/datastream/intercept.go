package datastream

import (
	"io"
	"sync"
)

type WriteCloserWithError = interface {
	io.WriteCloser
	CloseWithError(err error) error
}

// ComposableDataStreamIntercept is a structure for intercepting and manipulating data passed through a data stream
type ComposableDataStreamIntercept struct {
	errorHandler func(error)
	mutex        *sync.RWMutex
	reader       io.Reader
	writer       WriteCloserWithError
}

func newComposableDataStreamIntercept(reader io.Reader, writer WriteCloserWithError, errorHandler func(error)) *ComposableDataStreamIntercept {
	return &ComposableDataStreamIntercept{
		errorHandler: errorHandler,
		mutex:        &sync.RWMutex{},
		reader:       reader,
		writer:       writer,
	}
}

func (intercept *ComposableDataStreamIntercept) Read(p []byte) (int, error) {
	intercept.mutex.RLock()
	defer intercept.mutex.RUnlock()
	return intercept.reader.Read(p)
}

func (intercept *ComposableDataStreamIntercept) Write(p []byte) (int, error) {
	intercept.mutex.Lock()
	defer intercept.mutex.Unlock()
	return intercept.writer.Write(p)
}

// Close closes the data stream intercept to indicate that all data has been written
func (intercept *ComposableDataStreamIntercept) Close() error {
	intercept.mutex.Lock()
	defer intercept.mutex.Unlock()
	return intercept.writer.Close()
}

func (intercept *ComposableDataStreamIntercept) CloseWithError(err error) error {
	intercept.mutex.Lock()
	defer intercept.mutex.Unlock()
	return intercept.writer.CloseWithError(err)
}
