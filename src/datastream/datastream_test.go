package datastream

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComposableDataStream_CloseAlreadyClosed(t *testing.T) {
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Fail(t, "unexpected error")
	})

	require.False(t, stream.Closed())

	stream.Close()
	stream.Close()
	stream.Close()

	require.True(t, stream.Closed())
}

func TestComposableDataStream_Bytes(t *testing.T) {
	stream := NewComposableDataStream("name", func(caughtErr error) {
		require.Fail(t, "unexpected error")
	})

	stream.Replace(bytes.NewReader([]byte{1, 2, 3, 4}))
	stream.Close()
	stream.Wait()

	require.Equal(t, []byte{1, 2, 3, 4}, stream.Bytes())
}

func TestComposableDataStream_NewClosedComposableDataStreamFromBuffer(t *testing.T) {
	buffer := new(bytes.Buffer)
	buffer.Write([]byte("test 123"))
	stream := NewClosedComposableDataStreamFromBuffer(buffer)

	stream.Wait()
	require.Equal(t, []byte("test 123"), stream.Bytes())
}

func TestComposableDataStream_Len(t *testing.T) {
	buffer := new(bytes.Buffer)
	buffer.Write([]byte("test 123"))
	stream := NewClosedComposableDataStreamFromBuffer(buffer)

	stream.Wait()
	require.Equal(t, 8, stream.Len())
}
