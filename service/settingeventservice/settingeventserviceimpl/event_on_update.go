package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
)

func (s *service) OnUpdate(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.UpdateEvent,
) (err error) {
	return nil
}
