package io

import (
	"io"
	"io/ioutil"
	"sync"
)

// PipedWriteCloser is a simple struct that wraps an io.Pipe and a sync.WaitGroup for convenience
//
// Use a PipedWriteCloser when a WriteCloser is expected by some interface
// and you need a simple way of reading all data written once the io.WriteCloser has been closed
type PipedWriteCloser struct {
	reader    io.Reader
	result    []byte
	waitGroup *sync.WaitGroup
	writer    io.WriteCloser
}

// NewPipedWriteCloser creates a new PipedWriteCloser
//
// The PipedWriteCloser will be ready to be written to
// and the sync.WaitGroup will be waiting for the io.WriteCloser to close,
// after which it will store the result and unblock any previous calls to Wait()
func NewPipedWriteCloser() *PipedWriteCloser {
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	writeCloser := PipedWriteCloser{
		reader:    reader,
		waitGroup: waitGroup,
		writer:    writer,
	}
	go func() {
		writeCloser.result, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	return &writeCloser
}

// Writer is the interface that wraps the basic Write method.
func (writeCloser *PipedWriteCloser) Write(p []byte) (int, error) {
	return writeCloser.writer.Write(p)
}

// Closer is the interface that wraps the basic Close method.
func (writeCloser *PipedWriteCloser) Close() error {
	return writeCloser.writer.Close()
}

// Wait blocks until the WaitGroup counter is zero.
func (writeCloser *PipedWriteCloser) Wait() {
	writeCloser.waitGroup.Wait()
}

// Bytes returns all the data written to the writer as a slice of byte
//
// If called before Wait(), the result is undefined
func (writeCloser *PipedWriteCloser) Bytes() []byte {
	return writeCloser.result
}

// Bytes returns all the data written to the writer as a string
//
// If called before Wait(), the result is undefined
func (writeCloser *PipedWriteCloser) String() string {
	return string(writeCloser.result)
}
