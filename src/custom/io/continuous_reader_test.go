package io

import (
	"fmt"
	"github.com/stretchr/testify/require"
	systemio "io"
	"io/ioutil"
	"testing"
)

func TestContinuousReader_Read(t *testing.T) {
	testReader := NewTestReader(
		TestRead{chunk: []byte("test1\n"), err: nil},
		TestRead{chunk: []byte{}, err: systemio.EOF},
		TestRead{chunk: []byte("test2\n"), err: systemio.EOF},
		TestRead{chunk: []byte("test3\n"), err: fmt.Errorf("non EOF error")},
		TestRead{chunk: []byte("test4\n"), err: nil},
	)
	continuousReader := NewContinuousReader(testReader)
	completeOutput, outputErr := ioutil.ReadAll(continuousReader)
	require.Equal(t, "test1\ntest2\ntest3\n", string(completeOutput))
	require.NotNil(t, outputErr)
	require.Equal(t, "non EOF error", outputErr.Error())
}

type TestRead struct {
	chunk []byte
	err   error
}

type TestReader struct {
	reads []TestRead
}

func NewTestReader(reads ...TestRead) *TestReader {
	return &TestReader{
		reads: reads,
	}
}

func (testWriter *TestReader) Read(p []byte) (int, error) {
	var read TestRead
	read, testWriter.reads = testWriter.reads[0], testWriter.reads[1:]
	return copy(p, read.chunk), read.err
}
