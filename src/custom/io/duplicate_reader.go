package io

import (
	systemio "io"
)

// DuplicateReader copies reads from a single io.Reader into a pair of io.Readers
//
// It is the io.Reader analogue of an io.MultiWriter
func DuplicateReader(reader systemio.Reader, errorCallback func(error)) (systemio.Reader, systemio.Reader) {
	firstReader, firstWriter := systemio.Pipe()
	secondReader, secondWriter := systemio.Pipe()
	multiWriter := systemio.MultiWriter(firstWriter, secondWriter)
	go func() {
		_, err := systemio.Copy(multiWriter, reader)
		if err != nil {
			errorCallback(err)
			_ = firstWriter.CloseWithError(err)
			_ = secondWriter.CloseWithError(err)
		} else {
			_ = firstWriter.Close()
			_ = secondWriter.Close()
		}
	}()
	return firstReader, secondReader
}
