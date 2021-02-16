package io

import (
	"github.com/stretchr/testify/require"
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
	reader := NewErrorReader()
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
