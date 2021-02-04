package dir

import (
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
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
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	_ *middleware.ExecutionContext,
) {
	dirArgument := ""
	middleware.ParseArgumentsIncludingParents(&dirArgument, "dir", run)

	if dirArgument != "" {
		run.Log.DebugWithFields(
			log_fields.Symbol("📂"),
			log_fields.Message(dirArgument),
			log_fields.Middleware(dirMiddleware),
		)
		dirMiddleware.changeDirectory(dirArgument, run)
	}

	next(run)

	// now change back to initial working directory to make relative links work for next run
	dirMiddleware.changeDirectory(dirMiddleware.WorkingDir, run)
}

func (dirMiddleware DirMiddleware) changeDirectory(directory interface{}, run *models.PipelineRun) {
	dir, ok := directory.(string)
	if ok && dir != "" {
		err := dirMiddleware.DirChanger(dir)
		if err != nil {
			run.Log.Error(err)
		}
	}
}
