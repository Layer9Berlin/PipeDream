// Package stack defines a list of middleware items to be executed when a pipeline is run
package stack

import (
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/middleware/catch"
	"github.com/Layer9Berlin/pipedream/src/middleware/dir"
	"github.com/Layer9Berlin/pipedream/src/middleware/docker"
	"github.com/Layer9Berlin/pipedream/src/middleware/each"
	"github.com/Layer9Berlin/pipedream/src/middleware/env"
	"github.com/Layer9Berlin/pipedream/src/middleware/inherit"
	_input "github.com/Layer9Berlin/pipedream/src/middleware/input"
	"github.com/Layer9Berlin/pipedream/src/middleware/interpolate"
	_output "github.com/Layer9Berlin/pipedream/src/middleware/output"
	"github.com/Layer9Berlin/pipedream/src/middleware/pipe"
	_select "github.com/Layer9Berlin/pipedream/src/middleware/select"
	"github.com/Layer9Berlin/pipedream/src/middleware/shell"
	"github.com/Layer9Berlin/pipedream/src/middleware/ssh"
	_switch "github.com/Layer9Berlin/pipedream/src/middleware/switch"
	"github.com/Layer9Berlin/pipedream/src/middleware/timer"
	_waitfor "github.com/Layer9Berlin/pipedream/src/middleware/wait_for"
	"github.com/Layer9Berlin/pipedream/src/middleware/when"
)

// SetUpMiddleware returns the stack of middleware items that will be unwound during the run's execution
func SetUpMiddleware() []middleware.Middleware {
	return []middleware.Middleware{
		_select.NewMiddleware(),
		timer.NewMiddleware(),
		inherit.NewMiddleware(),
		_waitfor.NewMiddleware(),
		interpolate.NewMiddleware(),
		env.NewMiddleware(),
		when.NewMiddleware(),
		_output.NewMiddleware(),
		_switch.NewMiddleware(),
		catch.NewMiddleware(),
		pipe.NewMiddleware(),
		each.NewMiddleware(),
		shell.NewMiddleware(),
		ssh.NewMiddleware(),
		docker.NewMiddleware(),
		dir.NewMiddleware(),
		_input.NewMiddleware(),
	}
}
