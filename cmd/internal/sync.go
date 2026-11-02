package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/service/syncservice"
)

func InitSync(
	lc fx.Lifecycle,
	syncService syncservice.Service,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("checking the sync store...")
			return syncService.Start(ctx) //nolint:wrapcheck // the service's own
		},
		OnStop: func(_ context.Context) error {
			return syncService.Stop() //nolint:wrapcheck // the service's own
		},
	})
}
