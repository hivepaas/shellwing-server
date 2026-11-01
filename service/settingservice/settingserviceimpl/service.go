package settingserviceimpl

import (
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/settingservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func New(
	settingRepo repository.SettingRepo,

	userService userservice.Service,

	permissionManager permission.Manager,
) settingservice.Service {
	return &service{
		settingRepo: settingRepo,

		userService: userService,

		permissionManager: permissionManager,
	}
}

type service struct {
	settingRepo repository.SettingRepo

	userService userservice.Service

	permissionManager permission.Manager
}
