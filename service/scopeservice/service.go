package scopeservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	LoadObjectScope(ctx context.Context, db database.IDB, scopeType base.ObjectScopeType,
		objectID string, requireActive bool) (*entity.ObjectScope, error)
	LoadObjectScopeData(ctx context.Context, db database.IDB, scope *entity.ObjectScope) error
}
