package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

func TestValidateConfig(t *testing.T) {
	logger := logging.GlobalLogger()

	t.Run("outside development the app needs both secrets", func(t *testing.T) {
		cfg := &config.Config{Env: config.EnvBeta, RunMode: config.RunModeApp}
		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)

		cfg.Session.JWTSecret = "0123456789abcdef0123456789abcdef"
		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)

		cfg.Secret = "0123456789abcdef"
		assert.NoError(t, validateConfig(cfg, logger))
	})
}
