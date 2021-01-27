package logging

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/vbauerster/mpb/v5"
	"io"
	"os"
	"strings"
	"sync"
)

var simpleActivityIndicatorWidth = 1

type SimpleActivityIndicator struct {
	waitGroup *sync.WaitGroup

	len     int
	writer  io.Writer
	Subject ActivityIndicatingSubject
	visible bool
}

func NewSimpleActivityIndicator(writer io.Writer, options ...mpb.ContainerOption) *SimpleActivityIndicator {
	waitGroup := &sync.WaitGroup{}
	options = append(options, mpb.WithWaitGroup(waitGroup))
	options = append(options, mpb.WithWidth(nestedActivityIndicatorWidth))

	indicator := SimpleActivityIndicator{
		len:       0,
		visible:   true,
		waitGroup: waitGroup,
		writer:    writer,
	}
	return &indicator
}

func (activityIndicator *SimpleActivityIndicator) AddIndicator(subject ActivityIndicatingSubject, indentation int) {
	if !activityIndicator.visible {
		return
	}
	_, _ = activityIndicator.writer.Write([]byte(fmt.Sprint(strings.Repeat(" ", indentation), aurora.Blue("▶"), " ", subject, "\n")))
}

func (activityIndicator *SimpleActivityIndicator) Len() int {
	return activityIndicator.len
}

func (activityIndicator *SimpleActivityIndicator) Wait() {
	activityIndicator.waitGroup.Wait()
	activityIndicator.visible = false
	_ = os.Stdout.Sync()
	_ = os.Stderr.Sync()
}

func (activityIndicator *SimpleActivityIndicator) cancel() {
	activityIndicator.visible = false
}

func (activityIndicator *SimpleActivityIndicator) SetVisible(visible bool) {
	if activityIndicator.visible && !visible {
		activityIndicator.cancel()
	}
}

func (activityIndicator *SimpleActivityIndicator) Visible() bool {
	return activityIndicator.visible
}
