package notificationserviceimpl

import (
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/notificationservice"
	"github.com/hivepaas/shellwing-server/service/settingservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func New(
	settingRepo repository.SettingRepo,

	settingService settingservice.Service,
	userService userservice.Service,
) notificationservice.Service {
	return &service{
		settingRepo: settingRepo,

		settingService: settingService,
		userService:    userService,
	}
}

type service struct {
	settingRepo repository.SettingRepo

	settingService settingservice.Service
	userService    userservice.Service
}
