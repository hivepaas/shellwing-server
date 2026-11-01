package settinginitservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	InitDefaults(ctx context.Context, db database.IDB) error
	InitDefaultsWithTx(ctx context.Context, db database.Tx) error
}
