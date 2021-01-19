package custom_io

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestDuplicateReader(t *testing.T) {
	reader := strings.NewReader("test")
	errors := make([]error, 0, 10)
	reader1, reader2 := DuplicateReader(reader, func(err error) {
		errors = append(errors, err)
	})

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	go func() {
		defer testWaitGroup.Done()
		chunk, err := ioutil.ReadAll(reader1)
		require.Nil(t, err)
		require.Equal(t, "test", string(chunk))
	}()
	testWaitGroup.Add(1)
	go func() {
		defer testWaitGroup.Done()
		chunk, err := ioutil.ReadAll(reader2)
		require.Nil(t, err)
		require.Equal(t, "test", string(chunk))
	}()
	testWaitGroup.Wait()

	require.Equal(t, 0, len(errors))
}

func TestDuplicateReader_CloseError(t *testing.T) {
	reader := NewErrorReader(1)
	errors := make([]error, 0, 10)
	reader1, reader2 := DuplicateReader(reader, func(err error) {
		errors = append(errors, err)
	})

	testWaitGroup := &sync.WaitGroup{}
	testWaitGroup.Add(1)
	go func() {
		defer testWaitGroup.Done()
		chunk, err := ioutil.ReadAll(reader1)
		require.NotNil(t, err)
		require.Equal(t, "test error", err.Error())
		require.Equal(t, "", string(chunk))
	}()
	testWaitGroup.Add(1)
	go func() {
		defer testWaitGroup.Done()
		chunk, err := ioutil.ReadAll(reader2)
		require.NotNil(t, err)
		require.Equal(t, "test error", err.Error())
		require.Equal(t, "", string(chunk))
	}()
	testWaitGroup.Wait()

	require.Equal(t, 1, len(errors))
	require.Equal(t, "test error", errors[0].Error())
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