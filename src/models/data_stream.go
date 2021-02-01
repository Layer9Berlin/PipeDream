package models

import (
	"bytes"
	"io"
	"sync"
)

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
		err := stream.inputWriter.Close()
		if err != nil {
			stream.errorHandler(err)
		}
	}()
	go func() {
		defer stream.completionWaitGroup.Done()
		_, err := io.Copy(stream.result, stream.outputReader)
		if err != nil {
			stream.errorHandler(err)
		}
		stream.completed = true
		// the counter is initialized to 1,
		// so that the stream does not complete before it has been closed
	}()
}

func (stream *ComposableDataStream) Closed() bool {
	return stream.closed
}

func (stream *ComposableDataStream) Wait() {
	stream.completionWaitGroup.Wait()
}

func (stream *ComposableDataStream) Completed() bool {
	return stream.completed
}

func (stream *ComposableDataStream) String() string {
	return stream.result.String()
}

func (stream *ComposableDataStream) Bytes() []byte {
	return stream.result.Bytes()
}

func (stream *ComposableDataStream) Len() int {
	return stream.result.Len()
}
