package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

func InitDBConnection(lc fx.Lifecycle, db *database.DB, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("pinging db connection...")
			if err := db.Ping(); err != nil {
				logger.Errorf("failed to use connection %v", err.Error())

				return hperrors.Wrap(err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing db connection...")
			return db.Close()
		},
	})
}
