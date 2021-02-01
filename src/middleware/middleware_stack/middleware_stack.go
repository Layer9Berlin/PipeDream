package middleware_stack

import (
	"pipedream/src/middleware"
	"pipedream/src/middleware/catch"
	"pipedream/src/middleware/dir"
	"pipedream/src/middleware/docker"
	"pipedream/src/middleware/each"
	"pipedream/src/middleware/env"
	"pipedream/src/middleware/inherit"
	inputMiddleware "pipedream/src/middleware/input"
	"pipedream/src/middleware/interpolate"
	outputMiddleware "pipedream/src/middleware/output"
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
