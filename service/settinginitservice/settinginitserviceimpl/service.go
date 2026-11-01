package settinginitserviceimpl

import (
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/settinginitservice"
)

func New(
	lockRepo repository.LockRepo,
	settingRepo repository.SettingRepo,
) settinginitservice.Service {
	return &service{
		lockRepo:    lockRepo,
		settingRepo: settingRepo,
	}
}

type service struct {
	lockRepo    repository.LockRepo
	settingRepo repository.SettingRepo
}
