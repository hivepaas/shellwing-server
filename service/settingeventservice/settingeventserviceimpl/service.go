package settingeventserviceimpl

import (
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
)

func New(
	systemEventBus systemeventbusservice.Service,
) settingeventservice.Service {
	return &service{
		systemEventBus: systemEventBus,
	}
}

type service struct {
	systemEventBus systemeventbusservice.Service
}
