package settingeventservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	OnCreate(ctx context.Context, db database.IDB, event *CreateEvent) error
	OnUpdate(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnUpdateStatus(ctx context.Context, db database.IDB, event *UpdateEvent) error
	OnDelete(ctx context.Context, db database.IDB, event *DeleteEvent) error
}
