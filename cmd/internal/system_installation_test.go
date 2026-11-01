package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

type errorLogger struct {
	logging.Logger
	errors []string
}

func (l *errorLogger) Errorf(template string, args ...any) {
	l.errors = append(l.errors, fmt.Sprintf(template, args...))
}

func TestForgetFirstBootEnvRemovesTheFile(t *testing.T) {
	appPath := t.TempDir()
	t.Setenv("SHW_APP_PATH", appPath)
	assert.NoError(t, os.WriteFile(config.FirstBootEnvPath(appPath), []byte("SHW_USER_ADMIN_PASSWORD=x\n"), 0o600))
	logger := &errorLogger{}

	forgetFirstBootEnv(logger)

	_, err := os.Stat(config.FirstBootEnvPath(appPath))
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.Empty(t, logger.errors)
}

func TestForgetFirstBootEnvLogsWhatItCannotRemove(t *testing.T) {
	appPath := t.TempDir()
	t.Setenv("SHW_APP_PATH", appPath)
	// A directory with something in it is what os.Remove refuses.
	assert.NoError(t, os.MkdirAll(filepath.Join(config.FirstBootEnvPath(appPath), "x"), 0o700))
	logger := &errorLogger{}

	forgetFirstBootEnv(logger)

	assert.Len(t, logger.errors, 1)
}
