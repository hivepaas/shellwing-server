package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

func InitSystemSettings(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting system settings initialization...")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping system settings initialization...")
			return nil
		},
	})
}
