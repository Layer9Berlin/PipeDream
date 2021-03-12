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

func (syncBuffer *SynchronizedBuffer) Read(p []byte) (int, error) {
	syncBuffer.mutex.RLock()
	defer syncBuffer.mutex.RUnlock()
	return syncBuffer.buffer.Read(p)
}

func (syncBuffer *SynchronizedBuffer) Write(p []byte) (int, error) {
	syncBuffer.mutex.Lock()
	defer syncBuffer.mutex.Unlock()
	return syncBuffer.buffer.Write(p)
}

func (syncBuffer *SynchronizedBuffer) String() string {
	syncBuffer.mutex.RLock()
	defer syncBuffer.mutex.RUnlock()
	return syncBuffer.buffer.String()
}

func (syncBuffer *SynchronizedBuffer) Bytes() []byte {
	syncBuffer.mutex.RLock()
	defer syncBuffer.mutex.RUnlock()
	return syncBuffer.buffer.Bytes()
}
