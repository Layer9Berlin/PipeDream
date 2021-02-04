package save

import (
	"github.com/Layer9Berlin/pipedream/src/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestSaveMiddleware_Apply(t *testing.T) {
	run, _ := models.NewPipelineRun(nil, map[string]interface{}{
		"save": "ENV_VAR_NAME",
	}, nil, nil)

	run.Log.SetLevel(logrus.TraceLevel)
	values := make(map[string]string, 1)
	NewSaveMiddlewareWithValueSetter(func(key string, value string) error {
		values[key] = value
		return nil
	}).Apply(
		run,
		func(executingRun *models.PipelineRun) {
			executingRun.Stdout.Replace(strings.NewReader("test result"))
		},
		nil,
	)
	run.Close()
	run.Wait()

	require.Equal(t, 0, run.Log.ErrorCount())
	require.Equal(t, "test result", values["ENV_VAR_NAME"])
	require.Contains(t, run.Log.String(), "save")
}

func TestSaveMiddleware_NewSaveMiddleware(t *testing.T) {
	key := "TEST_PIPEDREAM_SAVEMIDDLEWARE_NEWSAVEMIDDLEWARE"
	_, haveValue := os.LookupEnv(key)
	require.False(t, haveValue)
	saveMiddleware := NewSaveMiddleware()
	err := saveMiddleware.valueSetter(key, "TEST")
	require.Nil(t, err)
	err = os.Unsetenv(key)
	require.Nil(t, err)
}
