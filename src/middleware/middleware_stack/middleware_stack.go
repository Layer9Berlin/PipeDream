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
	"pipedream/src/middleware/shell"
	"pipedream/src/middleware/ssh"
	syncMiddleware "pipedream/src/middleware/sync"
	"pipedream/src/middleware/timer"
	"pipedream/src/middleware/when"
)

func SetUpMiddleware() []middleware.Middleware {
	return []middleware.Middleware{
		timer.NewTimerMiddleware(),
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
	}
}
