package io

import (
	"bytes"
	"sync"
)

type SynchronizedBuffer struct {
	buffer *bytes.Buffer
	mutex  *sync.RWMutex
}

func NewSynchronizedBuffer() *SynchronizedBuffer {
	return &SynchronizedBuffer{
		buffer: new(bytes.Buffer),
		mutex:  &sync.RWMutex{},
	}
}

func (buffer *SynchronizedBuffer) Read(p []byte) (int, error) {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()
	return buffer.Read(p)
}

func (buffer *SynchronizedBuffer) Write(p []byte) (int, error) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	return buffer.Write(p)
}

func (buffer *SynchronizedBuffer) String() string {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()
	return buffer.String()
}

func (buffer *SynchronizedBuffer) Bytes() []byte {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()
	return buffer.Bytes()
}
