package models

import (
	"bytes"
	"fmt"
	"github.com/Layer9Berlin/pipedream/src/helpers/custom_io"
	"io"
	"io/ioutil"
)

func (stream *ComposableDataStream) CopyOrResult() io.Reader {
	if stream.closed {
		return bytes.NewReader(stream.result.Bytes())
	} else {
		firstReader, secondReader := custom_io.DuplicateReader(stream.outputReader, stream.errorHandler)
		stream.outputReader = firstReader
		return secondReader
	}
}

func (stream *ComposableDataStream) Copy() io.Reader {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	} else {
		firstReader, secondReader := custom_io.DuplicateReader(stream.outputReader, stream.errorHandler)
		stream.outputReader = firstReader
		return secondReader
	}
}

func (stream *ComposableDataStream) Intercept() io.ReadWriteCloser {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	} else {
		pipeReader, pipeWriter := io.Pipe()
		previousReader := stream.outputReader
		stream.outputReader = pipeReader
		return NewComposableDataStreamIntercept(previousReader, pipeWriter, stream.errorHandler)
	}
}

func (stream *ComposableDataStream) MergeWith(newReader io.Reader) {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
	} else {
		intercept := stream.Intercept()
		go func() {
			_, _ = io.Copy(intercept, intercept)
			_, err := io.Copy(intercept, newReader)
			if err != nil {
				stream.errorHandler(err)
			}
			_ = intercept.Close()
		}()
	}
}

func (stream *ComposableDataStream) Replace(reader io.Reader) {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
	} else {
		intercept := stream.Intercept()
		go func() {
			// be nice and read the previously defined input,
			// even though we discard the actual value
			// we might still want to unblock the corresponding writes
			_, _ = ioutil.ReadAll(intercept)
		}()
		go func() {
			_, err := io.Copy(intercept, reader)
			if err != nil {
				stream.errorHandler(err)
			}
			_ = intercept.Close()
		}()
	}
}

func (stream *ComposableDataStream) StartCopyingInto(writer io.Writer) {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
	} else {
		stream.outputReader = io.TeeReader(stream.outputReader, writer)
	}
}

func (stream *ComposableDataStream) WriteCloser() io.WriteCloser {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	} else {
		pipeReader, pipeWriter := io.Pipe()
		stream.MergeWith(pipeReader)
		return pipeWriter
	}
}
