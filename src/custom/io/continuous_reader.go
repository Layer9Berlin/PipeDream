package io

import (
	systemio "io"
)

func NewContinuousReader(reader systemio.Reader) systemio.Reader {
	return ContinuousReader{
		baseReader: reader,
	}
}

type ContinuousReader struct {
	baseReader systemio.Reader
}

func (reader ContinuousReader) Read(p []byte) (int, error) {
	n, err := reader.baseReader.Read(p)
	if err == systemio.EOF {
		err = nil
	}
	return n, err
}
