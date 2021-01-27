package middleware_stack

import (
	"pipedream/src/middleware"
	"pipedream/src/middleware/catch"
	"pipedream/src/middleware/dir"
	"pipedream/src/middleware/docker"
	"pipedream/src/middleware/each"
	"pipedream/src/middleware/env"
	"pipedream/src/middleware/inherit"
	"pipedream/src/middleware/interpolate"
	"pipedream/src/middleware/pipe"
	selectMiddleware "pipedream/src/middleware/select"
	"pipedream/src/middleware/shell"
	"pipedream/src/middleware/ssh"
	syncMiddleware "pipedream/src/middleware/sync"
	"pipedream/src/middleware/timer"
	"pipedream/src/middleware/wait"
	"pipedream/src/middleware/when"
)

func SetUpMiddleware() []middleware.Middleware {
	return []middleware.Middleware{
		selectMiddleware.NewSelectMiddleware(),
		timer.NewTimerMiddleware(),
		inherit.NewInheritMiddleware(),
		interpolate.NewInterpolateMiddleware(),
		env.NewEnvMiddleware(),
		catch.NewCatchMiddleware(),
		when.NewWhenMiddleware(),
		syncMiddleware.NewSyncMiddleware(),
		pipe.NewPipeMiddleware(),
		each.NewEachMiddleware(),
		wait.NewWaitMiddleware(),
		shell.NewShellMiddleware(),
		ssh.NewSshMiddleware(),
		docker.NewDockerMiddleware(),
		dir.NewDirMiddleware(),
	}
}
