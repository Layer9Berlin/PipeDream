package logging

import "fmt"

type ActivityIndicatingSubject interface {
	fmt.Stringer
	Wait()
	Completed() bool
}

type ActivityIndicator interface {
	AddIndicator(subject ActivityIndicatingSubject, indentation int)
	Len() int
	SetVisible(visible bool)
	Visible() bool
	Wait()
}
