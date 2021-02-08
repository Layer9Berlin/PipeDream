// Package dir provides a middleware for changing the current working directory
package dir

import (
	"github.com/Layer9Berlin/pipedream/src/logging/fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/pipeline"
	"os"
)

// Directory Navigator
type DirMiddleware struct {
	WorkingDir string
	DirChanger func(string) error
}

func (_ DirMiddleware) String() string {
	return "dir"
}

func NewDirMiddleware() DirMiddleware {
	workingDir, _ := os.Getwd()
	return DirMiddleware{
		DirChanger: os.Chdir,
		WorkingDir: workingDir,
	}
}

func (dirMiddleware DirMiddleware) Apply(
	run *pipeline.Run,
	next func(*pipeline.Run),
	_ *middleware.ExecutionContext,
) {
	dirArgument := ""
	pipeline.ParseArgumentsIncludingParents(&dirArgument, "dir", run)

	if dirArgument != "" {
		run.Log.DebugWithFields(
			fields.Symbol("📂"),
			fields.Message(dirArgument),
			fields.Middleware(dirMiddleware),
		)
		dirMiddleware.changeDirectory(dirArgument, run)
	}

	next(run)

	// now change back to initial working directory to make relative links work for next run
	dirMiddleware.changeDirectory(dirMiddleware.WorkingDir, run)
}

func (dirMiddleware DirMiddleware) changeDirectory(directory interface{}, run *pipeline.Run) {
	dir, ok := directory.(string)
	if ok && dir != "" {
		err := dirMiddleware.DirChanger(dir)
		if err != nil {
			run.Log.Error(err)
		}
	}
}
