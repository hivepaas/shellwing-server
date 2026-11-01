package emailserviceimpl

import (
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/emailservice"
)

func New(
	settingRepo repository.SettingRepo,
) emailservice.Service {
	return &service{
		settingRepo: settingRepo,
	}
}

type service struct {
	settingRepo repository.SettingRepo
}
