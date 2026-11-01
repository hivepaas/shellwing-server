package scopeserviceimpl

import (
	"github.com/hivepaas/shellwing-server/service/scopeservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func New(
	userService userservice.Service,
) scopeservice.Service {
	return &service{
		userService: userService,
	}
}

type service struct {
	userService userservice.Service
}
