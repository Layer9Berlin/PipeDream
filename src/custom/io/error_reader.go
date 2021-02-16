package io

import (
	"fmt"
)

type ErrorReader struct {
}

func NewErrorReader() *ErrorReader {
	return &ErrorReader{}
}

func (errorWriter *ErrorReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}
