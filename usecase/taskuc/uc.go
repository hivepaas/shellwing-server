package taskuc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/settingservice"
	"github.com/hivepaas/shellwing-server/service/taskservice"
)

type UC struct {
	db *database.DB

	settingRepo repository.SettingRepo

	auditService   auditservice.Service
	settingService settingservice.Service
	taskService    taskservice.Service
}

func New(
	db *database.DB,

	settingRepo repository.SettingRepo,

	auditService auditservice.Service,
	settingService settingservice.Service,
	taskService taskservice.Service,
) *UC {
	return &UC{
		db: db,

		settingRepo: settingRepo,

		auditService:   auditService,
		settingService: settingService,
		taskService:    taskService,
	}
}
