package middleware_stack

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRun_MiddlewareStack_setUpMiddleware(t *testing.T) {
	middlewareStack := SetUpMiddleware()
	middlewareStrings := make([]string, 0, 20)
	for _, middlewareItem := range middlewareStack {
		middlewareStrings = append(middlewareStrings, middlewareItem.String())
	}
	require.Contains(t, middlewareStrings, "timer")
	require.Contains(t, middlewareStrings, "inherit")
	require.Contains(t, middlewareStrings, "interpolate")
	require.Contains(t, middlewareStrings, "env")
	require.Contains(t, middlewareStrings, "catch")
	require.Contains(t, middlewareStrings, "when")
	require.Contains(t, middlewareStrings, "sync")
	require.Contains(t, middlewareStrings, "pipe")
	require.Contains(t, middlewareStrings, "each")
	require.Contains(t, middlewareStrings, "shell")
	require.Contains(t, middlewareStrings, "ssh")
	require.Contains(t, middlewareStrings, "docker")
	require.Contains(t, middlewareStrings, "dir")
}
