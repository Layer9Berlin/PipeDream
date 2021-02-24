package shell

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_StartClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.NotNil(t, commandExecutor.Start())
}

func Test_CallCmdStdinOnClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.Nil(t, commandExecutor.CmdStdin())
}

func Test_CallCmdStdoutOnClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.Nil(t, commandExecutor.CmdStdout())
}

func Test_CallCmdStderrOnClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.Nil(t, commandExecutor.CmdStderr())
}

func Test_KillClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.Nil(t, commandExecutor.Kill())
}

func Test_CallWaitOnClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.NotNil(t, commandExecutor.Wait())
}

func Test_CallStringOnClearedCommandExecutor(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	require.NotNil(t, commandExecutor.String())
}

func Test_KillCompletedCommand(t *testing.T) {
	commandExecutor := newDefaultCommandExecutor()
	commandExecutor.Init("echo", "'test'")
	_ = commandExecutor.Start()
	// wait to make test result reproducible
	_ = commandExecutor.Wait()
	require.NotNil(t, commandExecutor.Kill())
	require.Nil(t, commandExecutor.command)
}
