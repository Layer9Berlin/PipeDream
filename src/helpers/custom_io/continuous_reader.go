package custom_io

import (
	"io"
)

func NewContinuousReader(reader io.Reader) io.Reader {
	return ContinuousReader{
		baseReader: reader,
	}
}

type ContinuousReader struct {
	baseReader io.Reader
}

func (reader ContinuousReader) Read(p []byte) (int, error) {
	n, err := reader.baseReader.Read(p)
	if err == io.EOF {
		err = nil
	}
	return n, err
}
