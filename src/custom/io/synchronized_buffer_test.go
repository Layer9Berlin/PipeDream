package io

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"testing"
)

func TestSynchronizedBuffer(t *testing.T) {
	syncBuffer := NewSynchronizedBuffer()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			_, _ = syncBuffer.Write([]byte("test"))
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	require.Equal(t, strings.Repeat("test", 100), syncBuffer.String())
	require.Equal(t, []byte(strings.Repeat("test", 100)), syncBuffer.Bytes())
}

func TestSynchronizedBuffer_Read(t *testing.T) {
	syncBuffer := NewSynchronizedBuffer()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(200)
	result := new(bytes.Buffer)
	mutex := &sync.Mutex{}
	for i := 0; i < 100; i++ {
		go func() {
			_, _ = syncBuffer.Write([]byte("test"))
			waitGroup.Done()
		}()
		go func() {
			mutex.Lock()
			defer mutex.Unlock()
			buffer := make([]byte, 1024)
			n, _ := syncBuffer.Read(buffer)
			result.Write(buffer[:n])
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	buffer := make([]byte, 1024)
	n, _ := syncBuffer.Read(buffer)
	mutex.Lock()
	defer mutex.Unlock()
	result.Write(buffer[:n])
	require.Equal(t, strings.Repeat("test", 100), result.String())
}
