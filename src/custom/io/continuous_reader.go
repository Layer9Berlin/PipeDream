package io

import (
	systemio "io"
)

// ContinuousReader is an io.Reader that wraps another reader and ignores io.EOF markers
//
// This can be useful if you want to continue reading from a buffered reader
// after the buffer has been exhausted, for example
type ContinuousReader struct {
	baseReader systemio.Reader
}

// NewContinuousReader creates a new ContinuousReader by wrapping an existing io.Reader
func NewContinuousReader(reader systemio.Reader) systemio.Reader {
	return ContinuousReader{
		baseReader: reader,
	}
}

// NewContinuousReader creates a new ContinuousReader by wrapping an existing io.Reader
//
// Read reads up to len(p) bytes into p. If an io.EOF error is returned
// by the underlying reader, it is silently ignored.
func (reader ContinuousReader) Read(p []byte) (int, error) {
	n, err := reader.baseReader.Read(p)
	if err == systemio.EOF {
		err = nil
	}
	return n, err
}
