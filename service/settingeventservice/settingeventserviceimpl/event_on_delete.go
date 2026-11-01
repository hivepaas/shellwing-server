package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
)

func (s *service) OnDelete(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.DeleteEvent,
) (err error) {
	return nil
}
