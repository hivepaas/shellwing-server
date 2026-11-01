package startupserviceimpl

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/startupservice"
)

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,
) startupservice.Service {
	return &service{
		db: db,

		settingRepo: settingRepo,
	}
}

type service struct {
	db *database.DB

	settingRepo repository.SettingRepo
}
