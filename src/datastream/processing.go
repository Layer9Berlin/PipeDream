package datastream

import (
	"fmt"
	customio "github.com/Layer9Berlin/pipedream/src/custom/io"
	"io"
	"io/ioutil"
)

// Copy provides all data passing through the data stream through the io.Reader interface
//
// This function should not be called on closed data streams.
// You should read all data until io.EOF to prevent deadlocks.
func (stream *ComposableDataStream) Copy() io.Reader {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	}

	firstReader, secondReader := customio.DuplicateReader(stream.outputReader, stream.errorHandler)
	stream.outputReader = firstReader
	return secondReader
}

// Intercept provides a way of reading and manipulating all data passing through the data stream
//
// The data stream's input is available through the intercept's io.Reader interface.
// You should read all input data provided in this way (until io.EOF) to prevent deadlocks.
// The intercepted data stream's output will be replaced completely by the data written to the io.Writer interface.
// You must close the data stream when you have written all output data through the io.Writer.
func (stream *ComposableDataStream) Intercept() io.ReadWriteCloser {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	}

	pipeReader, pipeWriter := io.Pipe()
	previousReader := stream.outputReader
	stream.outputReader = pipeReader
	return newComposableDataStreamIntercept(previousReader, pipeWriter, stream.errorHandler)
}

// MergeWith combines the data passing through the data stream with the data provided by an io.Reader
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

// Replace replaces the data passing through the data stream with the data provided by an io.Reader
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

// StartCopyingInto writes the data passing through the data stream into the provided io.Writer
//
// Call Wait on the data stream to ensure that the writes have completed.
func (stream *ComposableDataStream) StartCopyingInto(writer io.Writer) {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
	} else {
		stream.outputReader = io.TeeReader(stream.outputReader, writer)
	}
}

// WriteCloser creates an io.WriteCloser that allows for additional data to be written into the data stream
//
// The data will be appended to the end of any existing data that might be passing through the data stream.
// You must close the io.WriteCloser once you have finished writing to it.
func (stream *ComposableDataStream) WriteCloser() io.WriteCloser {
	if stream.closed {
		stream.errorHandler(fmt.Errorf("attempt to modify closed data stream %q", stream.Name))
		return nil
	}

	pipeReader, pipeWriter := io.Pipe()
	stream.MergeWith(pipeReader)
	return pipeWriter
}
