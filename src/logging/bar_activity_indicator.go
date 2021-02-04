package logging

import (
	"github.com/vbauerster/mpb/v5"
	"io"
	"os"
	"sync"
)

type BarActivityIndicator struct {
	progress  *mpb.Progress
	waitGroup *sync.WaitGroup
	bar       *mpb.Bar

	len     int
	writer  io.Writer
	Subject ActivityIndicatingSubject
	visible bool
}

func NewBarActivityIndicator(writer io.Writer, options ...mpb.ContainerOption) *BarActivityIndicator {
	waitGroup := &sync.WaitGroup{}
	options = append(options, mpb.WithWaitGroup(waitGroup))
	options = append(options, mpb.WithWidth(64))

	progress := mpb.New(options...)
	indicator := BarActivityIndicator{
		bar:       nil,
		len:       0,
		progress:  progress,
		visible:   true,
		waitGroup: waitGroup,
		writer:    writer,
	}
	return &indicator
}

func (activityIndicator *BarActivityIndicator) AddIndicator(subject ActivityIndicatingSubject, _ int) {
	if !activityIndicator.visible {
		return
	}
	if activityIndicator.bar == nil {
		activityIndicator.bar = activityIndicator.progress.AddBar(1)
	} else {
		activityIndicator.bar.SetTotal(int64(activityIndicator.len), false)
	}
	activityIndicator.waitGroup.Add(1)
	go func() {
		subject.Wait()
		activityIndicator.bar.SetCurrent(activityIndicator.bar.Current() + 1)
		activityIndicator.waitGroup.Done()
	}()
}

func (activityIndicator *BarActivityIndicator) Len() int {
	return activityIndicator.len
}

func (activityIndicator *BarActivityIndicator) Wait() {
	activityIndicator.waitGroup.Wait()
	activityIndicator.visible = false
	_ = os.Stdout.Sync()
	_ = os.Stderr.Sync()
}

func (activityIndicator *BarActivityIndicator) cancel() {
	activityIndicator.visible = false
	activityIndicator.bar.Abort(true)
}

func (activityIndicator *BarActivityIndicator) SetVisible(visible bool) {
	if activityIndicator.visible && !visible {
		activityIndicator.cancel()
	}
}

func (activityIndicator *BarActivityIndicator) Visible() bool {
	return activityIndicator.visible
}
