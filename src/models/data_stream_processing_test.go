package models

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	strings "strings"
	"sync"
	"testing"
)

func TestComposableDataStream_Close(t *testing.T) {
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Fail(t, "unexpected error")
	})
	stream.MergeWith(strings.NewReader("test"))

	require.False(t, stream.Closed())
	require.False(t, stream.Completed())

	stream.Close()

	require.True(t, stream.Closed())
	require.False(t, stream.Completed())

	stream.Wait()

	require.True(t, stream.Closed())
	require.True(t, stream.Completed())
}

func TestComposableDataStream_Copy(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	reader := stream.Copy()
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	go func() {
		previousData, _ := ioutil.ReadAll(previousIntercept)
		require.Equal(t, "", string(previousData))
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		result, err := ioutil.ReadAll(nextIntercept)
		require.Nil(t, err)
		require.Equal(t, "previous data\n", string(result))
		_, err = nextIntercept.Write([]byte("intercepted\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	var resultCopy []byte = nil
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		resultCopy, _ = ioutil.ReadAll(reader)
	}()
	stream.Close()
	stream.Wait()
	waitGroup.Wait()

	require.Equal(t, "intercepted\n", stream.String())
	require.Equal(t, "previous data\n", string(resultCopy))
}

func TestComposableDataStream_Copy_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	_ = stream.Copy()

	require.True(t, expectedErrorEncountered)
}

func TestComposableDataStream_Intercept(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	intercept := stream.Intercept()
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	waitGroup := &sync.WaitGroup{}
	go func() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, _ = ioutil.ReadAll(previousIntercept)
		}()
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			result, err := ioutil.ReadAll(nextIntercept)
			require.Nil(t, err)
			require.Equal(t, "intercepted data\n", string(result))
		}()
		_, err := nextIntercept.Write([]byte("next data\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	go func() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			result, err := ioutil.ReadAll(intercept)
			require.Nil(t, err)
			require.Equal(t, "previous data\n", string(result))
		}()
		_, err := intercept.Write([]byte("intercepted data\n"))
		require.Nil(t, err)
		err = intercept.Close()
		require.Nil(t, err)
	}()
	stream.Close()
	stream.Wait()
	waitGroup.Wait()

	require.Equal(t, "next data\n", stream.String())
}

func TestComposableDataStream_Intercept_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	_ = stream.Intercept()

	require.True(t, expectedErrorEncountered)
}

func TestComposableDataStream_InterceptOrder(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	firstIntercept := stream.Intercept()
	secondIntercept := stream.Intercept()
	thirdIntercept := stream.Intercept()
	fourthIntercept := stream.Intercept()
	go func() {
		firstResult, err := ioutil.ReadAll(firstIntercept)
		require.Nil(t, err)
		require.Equal(t, "", string(firstResult))
		_, err = firstIntercept.Write([]byte("first output"))
		require.Nil(t, err)
		require.Nil(t, firstIntercept.Close())
	}()
	go func() {
		secondResult, err := ioutil.ReadAll(secondIntercept)
		require.Nil(t, err)
		require.Equal(t, "first output", string(secondResult))
		_, err = secondIntercept.Write([]byte("second output"))
		require.Nil(t, err)
		require.Nil(t, secondIntercept.Close())
	}()
	go func() {
		thirdResult, err := ioutil.ReadAll(thirdIntercept)
		require.Nil(t, err)
		require.Equal(t, "second output", string(thirdResult))
		_, err = thirdIntercept.Write([]byte("third output"))
		require.Nil(t, err)
		require.Nil(t, thirdIntercept.Close())
	}()
	go func() {
		fourthResult, err := ioutil.ReadAll(fourthIntercept)
		require.Nil(t, err)
		require.Equal(t, "third output", string(fourthResult))
		_, err = fourthIntercept.Write([]byte("fourth output"))
		require.Nil(t, err)
		require.Nil(t, fourthIntercept.Close())
	}()
	stream.Close()
	stream.Wait()

	require.Equal(t, "fourth output", stream.String())
}

func TestComposableDataStream_MergeWith(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	stream.MergeWith(strings.NewReader("test"))
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	go func() {
		previousData, _ := ioutil.ReadAll(previousIntercept)
		require.Equal(t, "", string(previousData))
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		result, err := ioutil.ReadAll(nextIntercept)
		require.Nil(t, err)
		require.Equal(t, "previous data\ntest", string(result))
		_, err = nextIntercept.Write([]byte("intercepted\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	stream.Close()
	stream.Wait()

	require.Equal(t, "intercepted\n", stream.String())
}

func TestComposableDataStream_MergeWith_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	stream.MergeWith(strings.NewReader("test"))

	require.True(t, expectedErrorEncountered)
}

func TestComposableDataStream_MergeWith_CloseError(t *testing.T) {
	var errorMessage = ""
	stream := NewComposableDataStream("name", func(caughtErr error) {
		if caughtErr != nil {
			errorMessage = caughtErr.Error()
		}
	})
	stream.MergeWith(NewErrorReader(1))
	stream.Close()
	stream.Wait()

	require.Equal(t, "test error", errorMessage)
}

func TestComposableDataStream_Replace(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	stream.Replace(strings.NewReader("test input data"))
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	go func() {
		previousData, _ := ioutil.ReadAll(previousIntercept)
		require.Equal(t, "", string(previousData))
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		result, err := ioutil.ReadAll(nextIntercept)
		require.Nil(t, err)
		require.Equal(t, "test input data", string(result))
		_, err = nextIntercept.Write([]byte("intercepted\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	stream.Close()
	stream.Wait()

	require.Equal(t, "intercepted\n", stream.String())
}

func TestComposableDataStream_Replace_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	stream.Replace(strings.NewReader("test"))

	require.True(t, expectedErrorEncountered)
}

func TestComposableDataStream_Replace_CloseError(t *testing.T) {
	var errorMessage = ""
	stream := NewComposableDataStream("name", func(caughtErr error) {
		if caughtErr != nil {
			errorMessage = caughtErr.Error()
		}
	})
	stream.Replace(NewErrorReader(1))
	stream.Close()
	stream.Wait()

	require.Equal(t, "test error", errorMessage)
}

func TestComposableDataStream_StartCopyingInto(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	buffer := new(bytes.Buffer)
	stream.StartCopyingInto(buffer)
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	go func() {
		previousData, _ := ioutil.ReadAll(previousIntercept)
		require.Equal(t, "", string(previousData))
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		result, err := ioutil.ReadAll(nextIntercept)
		require.Nil(t, err)
		require.Equal(t, "previous data\n", string(result))
		_, err = nextIntercept.Write([]byte("intercepted\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	stream.Close()
	stream.Wait()

	require.Equal(t, "intercepted\n", stream.String())
	require.Equal(t, "previous data\n", buffer.String())
}

func TestComposableDataStream_StartCopyingInto_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	buffer := new(bytes.Buffer)
	stream.StartCopyingInto(buffer)

	require.True(t, expectedErrorEncountered)
}

func TestComposableDataStream_WriteCloser(t *testing.T) {
	// set up composable data stream
	stream := NewComposableDataStream("name", func(err error) {
		require.Nil(t, err)
	})
	// we intercept to set up some data before the appender is called
	previousIntercept := stream.Intercept()
	writeCloser := stream.WriteCloser()
	// intercept again to ensure correct order
	nextIntercept := stream.Intercept()
	go func() {
		previousData, _ := ioutil.ReadAll(previousIntercept)
		require.Equal(t, "", string(previousData))
		_, err := previousIntercept.Write([]byte("previous data\n"))
		require.Nil(t, err)
		require.Nil(t, previousIntercept.Close())
	}()
	go func() {
		result, err := ioutil.ReadAll(nextIntercept)
		require.Nil(t, err)
		require.Equal(t, "previous data\nwrite closer data\n", string(result))
		_, err = nextIntercept.Write([]byte("intercepted\n"))
		require.Nil(t, err)
		require.Nil(t, nextIntercept.Close())
	}()
	go func() {
		_, _ = io.WriteString(writeCloser, "write closer data\n")
		_ = writeCloser.Close()
	}()
	stream.Close()
	stream.Wait()

	require.Equal(t, "intercepted\n", stream.String())
}

func TestComposableDataStream_WriteCloser_AfterClosure(t *testing.T) {
	expectedErrorEncountered := false
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Contains(t, caughtErr.Error(), "attempt to modify closed data stream")
		expectedErrorEncountered = true
	})

	stream.Close()
	_ = stream.WriteCloser()

	require.True(t, expectedErrorEncountered)
}

type ErrorReader struct {
	counter int
}

func NewErrorReader(counter int) *ErrorReader {
	return &ErrorReader{
		counter: counter,
	}
}

func (errorWriter *ErrorReader)Read(p []byte) (int, error) {
	if errorWriter.counter <= 0 {
		return 0, io.EOF
	}
	errorWriter.counter = errorWriter.counter - 1
	return 0, fmt.Errorf("test error")
}
