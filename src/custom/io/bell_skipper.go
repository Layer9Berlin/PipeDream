package io

import (
	systemio "io"
)

// BellSkipper is a custom WriteCloser that skips `Bell` characters
//
// Used to prevent promptui making sounds when a button is pressed
type BellSkipper struct {
	baseOutput systemio.WriteCloser
}

// NewBellSkipper creates a new BellSkipper WriteCloser
func NewBellSkipper(baseOutput systemio.WriteCloser) *BellSkipper {
	return &BellSkipper{baseOutput: baseOutput}
}

// Write writes len(p) bytes from p to the underlying data stream.
func (skipper *BellSkipper) Write(b []byte) (int, error) {
	const charBell = 7
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return skipper.baseOutput.Write(b)
}

// Close closes the data stream
func (skipper *BellSkipper) Close() error {
	return skipper.baseOutput.Close()
}
