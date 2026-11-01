package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
)

func (s *service) OnCreate(
	ctx context.Context,
	_ database.IDB,
	event *settingeventservice.CreateEvent,
) (err error) {
	return nil
}
