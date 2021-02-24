package io

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestPipedWriteCloser(t *testing.T) {
	pipedWriteCloser := NewPipedWriteCloser()
	go func() {
		time.Sleep(200)
		_, _ = io.WriteString(pipedWriteCloser, "test")
		_ = pipedWriteCloser.Close()
	}()
	pipedWriteCloser.Wait()
	require.Equal(t, []byte("test"), pipedWriteCloser.Bytes())
	require.Equal(t, "test", pipedWriteCloser.String())
}
