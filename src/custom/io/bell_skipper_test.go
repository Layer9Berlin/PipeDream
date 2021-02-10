package io

import (
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

func TestBellSkipper_Write(t *testing.T) {
	pipeReader, pipeWriter := io.Pipe()
	bellSkipper := NewBellSkipper(pipeWriter)
	buffer := make([]byte, 0, 10)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		buffer, _ = ioutil.ReadAll(pipeReader)
		waitGroup.Done()
	}()
	for i := 0; i < 10; i++ {
		_, _ = bellSkipper.Write([]byte{byte(i)})
	}
	_ = bellSkipper.Close()
	waitGroup.Wait()
	require.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 8, 9}, buffer)
}
