package logging

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"github.com/vbauerster/mpb/v5"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

type TestStringer struct {
	Complete  bool
	Value     string
	WaitGroup *sync.WaitGroup
}

func (t TestStringer) String() string {
	return t.Value
}

func (t TestStringer) Wait() {
	t.WaitGroup.Wait()
}

func (t TestStringer) Completed() bool {
	return t.Complete
}

func NewTestStringer(value string) *TestStringer {
	return &TestStringer{
		Complete: false,
		Value:    value,
		WaitGroup: &sync.WaitGroup{},
	}
}

func TestNestedActivityIndicator_Shown(t *testing.T) {
	// need to set this for correct output when printing to buffer
	nestedActivityIndicatorWidth = 64

	subject1 := NewTestStringer("test1")
	subject2 := NewTestStringer("test2")
	subject1.WaitGroup.Add(1)
	subject2.WaitGroup.Add(1)
	buffer := new(bytes.Buffer)
	activityIndicator := NewNestedActivityIndicator(mpb.WithOutput(buffer))
	activityIndicator.AddSpinner(subject1, 2)
	activityIndicator.AddSpinner(subject2, 4)

	time.Sleep(300 * time.Millisecond)

	require.Contains(t, buffer.String(), "test1")
	require.Contains(t, buffer.String(), "test2")

	subject2.WaitGroup.Done()

	time.Sleep(300 * time.Millisecond)

	require.Contains(t, buffer.String(), "test1")
	require.Contains(t, buffer.String(), "test2")
	require.Contains(t, buffer.String(), "✔")
	subject1.WaitGroup.Done()

	activityIndicator.Wait()
}

func TestNestedActivityIndicator_Len(t *testing.T) {
	buffer := new(bytes.Buffer)
	activityIndicator := NewNestedActivityIndicator(mpb.WithOutput(buffer))
	require.Equal(t, 0, activityIndicator.Len())

	subject1 := NewTestStringer("test1")
	subject2 := NewTestStringer("test2")
	subject1.WaitGroup.Add(1)
	defer subject1.WaitGroup.Done()
	subject2.WaitGroup.Add(1)
	defer subject2.WaitGroup.Done()
	activityIndicator.AddSpinner(subject1, 2)
	activityIndicator.AddSpinner(subject2, 4)
	require.Equal(t, 2, activityIndicator.Len())
}

func TestNestedActivityIndicator_SetVisible(t *testing.T) {
	nestedActivityIndicatorWidth = 64
	buffer := new(bytes.Buffer)
	activityIndicator := NewNestedActivityIndicator(mpb.WithOutput(buffer))
	require.Equal(t, 0, activityIndicator.Len())

	subject := NewTestStringer("test")
	subject.WaitGroup.Add(1)
	defer subject.WaitGroup.Done()
	activityIndicator.AddSpinner(subject, 2)
	time.Sleep(300 * time.Millisecond)
	require.Equal(t, 1, activityIndicator.Len())
	activityIndicator.SetVisible(false)
	result, err := ioutil.ReadAll(buffer)
	require.Nil(t, err)
	require.Contains(t, string(result), "test")
	time.Sleep(300 * time.Millisecond)
	result, err = ioutil.ReadAll(buffer)
	require.Nil(t, err)
	require.NotContains(t, string(result), "test")
	require.Equal(t, 0, activityIndicator.Len())

	activityIndicator.AddSpinner(subject, 2)
	require.Equal(t, 0, activityIndicator.Len())
}

func TestNestedActivityIndicator_Wait(t *testing.T) {
	nestedActivityIndicatorWidth = 64
	buffer := new(bytes.Buffer)
	activityIndicator := NewNestedActivityIndicator(mpb.WithOutput(buffer))
	require.Equal(t, 0, activityIndicator.Len())

	subject1 := NewTestStringer("test1")
	subject2 := NewTestStringer("test2")
	subject1.WaitGroup.Add(1)
	subject2.WaitGroup.Add(1)
	activityIndicator.AddSpinner(subject1, 2)
	activityIndicator.AddSpinner(subject2, 4)
	go func() {
		defer subject1.WaitGroup.Done()
		time.Sleep(300 * time.Millisecond)
	}()
	go func() {
		defer subject2.WaitGroup.Done()
		time.Sleep(200 * time.Millisecond)
	}()
	activityIndicator.Wait()
	result, err := ioutil.ReadAll(buffer)
	require.Nil(t, err)
	require.Contains(t, string(result), "test")
	time.Sleep(300 * time.Millisecond)
	result, err = ioutil.ReadAll(buffer)
	require.Nil(t, err)
	require.NotContains(t, string(result), "test")
	require.Equal(t, 0, activityIndicator.Len())

	activityIndicator.AddSpinner(subject1, 2)
	require.Equal(t, 0, activityIndicator.Len())
}