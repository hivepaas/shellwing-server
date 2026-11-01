package auditloguc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/settingservice"
)

type UC struct {
	db *database.DB

	auditLogRepo repository.AuditLogRepo

	settingService settingservice.Service
}

func New(
	db *database.DB,

	auditLogRepo repository.AuditLogRepo,

	settingService settingservice.Service,
) *UC {
	return &UC{
		db: db,

		auditLogRepo: auditLogRepo,

		settingService: settingService,
	}
}
