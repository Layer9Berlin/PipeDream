package custom_io

import (
	"io"
)

func DuplicateReader(reader io.Reader, errorCallback func(error)) (io.Reader, io.Reader) {
	firstReader, firstWriter := io.Pipe()
	secondReader, secondWriter := io.Pipe()
	multiWriter := io.MultiWriter(firstWriter, secondWriter)
	go func() {
		_, err := io.Copy(multiWriter, reader)
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
