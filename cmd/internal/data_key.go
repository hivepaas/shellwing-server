package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/service/datakeyservice"
)

// InitDataKey installs the key stored secrets are encrypted with, before anything
// can read or write one.
//
// Failing here is deliberate: an app secret that does not open the stored key
// would otherwise surface much later as settings that cannot be decrypted, one
// request at a time.
func InitDataKey(
	lc fx.Lifecycle,
	db *database.DB,
	dataKeyService datakeyservice.Service,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("loading data encryption key...")
			if err := dataKeyService.Load(ctx, db); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
}
