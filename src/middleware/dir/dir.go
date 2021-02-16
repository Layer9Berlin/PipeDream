// Package dir provides a middleware for changing the current working directory
package dir

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"os"
)

// Middleware is a directory navigator
type Middleware struct {
	WorkingDir string
	DirChanger func(string) error
}

// String is a human-readable description
func (Middleware) String() string {
	return "dir"
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware() Middleware {
	// @TODO: do we need to resolve symlinks here?
	workingDir, _ := os.Getwd()
	return Middleware{
		DirChanger: os.Chdir,
		WorkingDir: workingDir,
	}
}

// Apply is where the middleware's logic resides
//
// It adapts the run based on its slice of the run's arguments.
// It may also trigger side effects such as executing shell commands or full runs of other pipelines.
// When done, this function should call next in order to continue unwinding the stack.
func (dirMiddleware Middleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	dirArgument := ""
	pipeline.ParseArgumentsIncludingParents(&dirArgument, "dir", run)

	if dirArgument != "" {
		run.Log.Debug(
			fields.Symbol("📂"),
			fields.Message(dirArgument),
			fields.Middleware(dirMiddleware),
		)
		// @TODO: provide a way of resolving the path relative to the pipeline file's location
		dirMiddleware.changeDirectory(dirArgument, run)
	}

	next(run)

	// now change back to initial working directory to make relative links work for next run
	dirMiddleware.changeDirectory(dirMiddleware.WorkingDir, run)
}

func (dirMiddleware Middleware) changeDirectory(directory string, run *pipeline.Run) {
	err := dirMiddleware.DirChanger(directory)
	if err != nil {
		run.Log.Error(err)
	}
}
