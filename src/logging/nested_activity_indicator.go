package logging

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"os"
	"strings"
	"sync"
)

var nestedActivityIndicatorWidth = 1

type NestedActivityIndicator struct {
	progress  *mpb.Progress
	waitGroup *sync.WaitGroup

	bars    []*mpb.Bar
	Subject ActivityIndicatingSubject
	visible bool
}

func NewNestedActivityIndicator(options ...mpb.ContainerOption) *NestedActivityIndicator {
	waitGroup := &sync.WaitGroup{}
	options = append(options, mpb.WithWaitGroup(waitGroup))
	options = append(options, mpb.WithWidth(nestedActivityIndicatorWidth))

	progress := mpb.New(options...)
	indicator := NestedActivityIndicator{
		bars:      make([]*mpb.Bar, 0, 12),
		progress:  progress,
		visible:   true,
		waitGroup: waitGroup,
	}
	return &indicator
}

func defaultNestedActivityIndicatorOptions(subject ActivityIndicatingSubject, indentation int) []mpb.BarOption {
	subjectDescription := decor.Any(func(statistics decor.Statistics) string {
		return subject.String()
	})
	return []mpb.BarOption{
		mpb.BarFillerOnComplete(fmt.Sprint(aurora.Green("✔"))),
		mpb.PrependDecorators(
			decor.Name(fmt.Sprint(strings.Repeat(" ", indentation))),
		),
		mpb.AppendDecorators(
			subjectDescription,
		),
		mpb.TrimSpace(),
	}
}

func (activityIndicator *NestedActivityIndicator) AddIndicator(subject ActivityIndicatingSubject, indentation int) {
	if !activityIndicator.visible {
		return
	}
	bar := activityIndicator.progress.AddSpinner(
		int64(1),
		mpb.SpinnerOnLeft,
		defaultNestedActivityIndicatorOptions(subject, indentation)...,
	)
	activityIndicator.bars = append(activityIndicator.bars, bar)
	activityIndicator.waitGroup.Add(1)
	go func() {
		defer activityIndicator.waitGroup.Done()
		subject.Wait()
		bar.Increment()
	}()
}

func (activityIndicator *NestedActivityIndicator) Len() int {
	return activityIndicator.progress.BarCount()
}

func (activityIndicator *NestedActivityIndicator) Wait() {
	activityIndicator.progress.Wait()
	activityIndicator.visible = false
	_ = os.Stdout.Sync()
	_ = os.Stderr.Sync()
}

func (activityIndicator *NestedActivityIndicator) cancel() {
	activityIndicator.visible = false
	for _, bar := range activityIndicator.bars {
		bar.Abort(true)
	}
}

func (activityIndicator *NestedActivityIndicator) SetVisible(visible bool) {
	if activityIndicator.visible && !visible {
		activityIndicator.cancel()
	}
}

func (activityIndicator *NestedActivityIndicator) Visible() bool {
	return activityIndicator.visible
}
