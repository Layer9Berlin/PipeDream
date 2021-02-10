// Package datastream provides a composable data stream, which pipes multiple inputs to multiple outputs according to the defined compositions
package datastream

import (
	"bytes"
	"io"
	"sync"
)

// ComposableDataStream is a data stream that can be composed with other data streams
//
// You can combine an arbitrary number of io.WriteCloser inputs and io.Reader outputs.
// Call Close when all compositions have been defined. This will start the data flow.
// Once all data has been passed through (all inputs are closed), the data stream will complete.
type ComposableDataStream struct {
	errorHandler func(error)
	inputWriter  io.WriteCloser
	Name         string
	outputReader io.Reader

	completed           bool
	completionMutex     *sync.Mutex
	completionWaitGroup *sync.WaitGroup
	finalizationMutex   *sync.Mutex
	closed              bool
	result              *bytes.Buffer
}

// NewComposableDataStream creates a new ComposableDataStream
//
// name is a description of the data stream for debugging convenience
func NewComposableDataStream(name string, errorHandler func(error)) *ComposableDataStream {
	reader, writer := io.Pipe()
	completionWaitGroup := &sync.WaitGroup{}
	// do not complete before Close() is called
	completionWaitGroup.Add(1)
	return &ComposableDataStream{
		// errors may occur in goroutines after function calls return,
		// so we define a handler that can be executed at any time
		errorHandler: errorHandler,
		inputWriter:  writer,
		Name:         name,
		outputReader: reader,

		completed:           false,
		completionMutex:     &sync.Mutex{},
		completionWaitGroup: completionWaitGroup,
		finalizationMutex:   &sync.Mutex{},
		closed:              false,
		result:              new(bytes.Buffer),
	}
}

// NewClosedComposableDataStreamFromBuffer creates an already closed ComposableDataStream with the specified result
func NewClosedComposableDataStreamFromBuffer(buffer *bytes.Buffer) *ComposableDataStream {
	completionWaitGroup := &sync.WaitGroup{}
	return &ComposableDataStream{
		completed:           true,
		completionMutex:     &sync.Mutex{},
		completionWaitGroup: completionWaitGroup,
		finalizationMutex:   &sync.Mutex{},
		closed:              true,
		result:              buffer,
	}
}

// Close finalizes the data stream
//
// It should be called when all compositions have been defined
// in order to start the data flow. It is a prerequisite
// for the data stream to complete.
func (stream *ComposableDataStream) Close() {
	stream.finalizationMutex.Lock()
	defer stream.finalizationMutex.Unlock()
	if stream.closed {
		return
	}
	stream.closed = true

	// we don't expect any more inputs
	go func() {
		// close the writer asynchronously, so that the reader knows we're done
		_ = stream.inputWriter.Close()
	}()
	go func() {
		defer stream.completionWaitGroup.Done()
		_, _ = io.Copy(stream.result, stream.outputReader)
		stream.completed = true
		// the counter is initialized to 1,
		// so that the stream does not complete before it has been closed
	}()
}

// Closed indicates whether the data stream has been closed
func (stream *ComposableDataStream) Closed() bool {
	return stream.closed
}

// Wait pauses execution until the data stream has completed
func (stream *ComposableDataStream) Wait() {
	stream.completionWaitGroup.Wait()
}

// Completed indicates whether the data stream has completed
func (stream *ComposableDataStream) Completed() bool {
	return stream.completed
}

// String returns the result of a completed data stream as a string
//
// If you call String before the data stream has completed, the result is undefined.
func (stream *ComposableDataStream) String() string {
	return stream.result.String()
}

// Bytes returns the result of a completed data stream as a byte slice
//
// If you call Bytes before the data stream has completed, the result is undefined.
func (stream *ComposableDataStream) Bytes() []byte {
	return stream.result.Bytes()
}

// Len returns the size of the result of a completed data stream in number of bytes
//
// If you call Len before the data stream has completed, the result is undefined.
func (stream *ComposableDataStream) Len() int {
	return stream.result.Len()
}
