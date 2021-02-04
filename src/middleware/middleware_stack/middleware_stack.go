package middleware_stack

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/middleware/catch"
	"github.com/Layer9Berlin/pipedream/src/middleware/dir"
	"github.com/Layer9Berlin/pipedream/src/middleware/docker"
	"github.com/Layer9Berlin/pipedream/src/middleware/each"
	"github.com/Layer9Berlin/pipedream/src/middleware/env"
	"github.com/Layer9Berlin/pipedream/src/middleware/inherit"
	inputMiddleware "github.com/Layer9Berlin/pipedream/src/middleware/input"
	"github.com/Layer9Berlin/pipedream/src/middleware/interpolate"
	outputMiddleware "github.com/Layer9Berlin/pipedream/src/middleware/output"
	"github.com/Layer9Berlin/pipedream/src/middleware/pipe"
	selectMiddleware "github.com/Layer9Berlin/pipedream/src/middleware/select"
	"github.com/Layer9Berlin/pipedream/src/middleware/shell"
	"github.com/Layer9Berlin/pipedream/src/middleware/ssh"
	syncMiddleware "github.com/Layer9Berlin/pipedream/src/middleware/sync"
	"github.com/Layer9Berlin/pipedream/src/middleware/timer"
	"github.com/Layer9Berlin/pipedream/src/middleware/wait"
	"github.com/Layer9Berlin/pipedream/src/middleware/when"
)

func SetUpMiddleware() []middleware.Middleware {
	return []middleware.Middleware{
		selectMiddleware.NewSelectMiddleware(),
		timer.NewTimerMiddleware(),
		wait.NewWaitMiddleware(),
		inherit.NewInheritMiddleware(),
		interpolate.NewInterpolateMiddleware(),
		env.NewEnvMiddleware(),
		catch.NewCatchMiddleware(),
		when.NewWhenMiddleware(),
		syncMiddleware.NewSyncMiddleware(),
		pipe.NewPipeMiddleware(),
		each.NewEachMiddleware(),
		shell.NewShellMiddleware(),
		ssh.NewSshMiddleware(),
		docker.NewDockerMiddleware(),
		dir.NewDirMiddleware(),
		inputMiddleware.NewInputMiddleware(),
		outputMiddleware.NewOutputMiddleware(),
	}
}
