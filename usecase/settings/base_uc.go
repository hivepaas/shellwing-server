package settings

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/scopeservice"
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
	"github.com/hivepaas/shellwing-server/service/settinginitservice"
	"github.com/hivepaas/shellwing-server/service/settingservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

type BaseUC struct {
	DB *database.DB

	FileRepo    repository.FileRepo
	ResLinkRepo repository.ResLinkRepo
	SettingRepo repository.SettingRepo
	TagRepo     repository.TagRepo

	AuditService        auditservice.Service
	ScopeService        scopeservice.Service
	SettingEventService settingeventservice.Service
	SettingInitService  settinginitservice.Service
	SettingService      settingservice.Service
	UserService         userservice.Service
	PermissionManager   permission.Manager
}

func New(
	db *database.DB,

	fileRepo repository.FileRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	tagRepo repository.TagRepo,

	auditService auditservice.Service,
	scopeService scopeservice.Service,
	settingEventService settingeventservice.Service,
	settingInitService settinginitservice.Service,
	settingService settingservice.Service,
	userService userservice.Service,
	permissionManager permission.Manager,
) *BaseUC {
	return &BaseUC{
		DB: db,

		FileRepo:    fileRepo,
		ResLinkRepo: resLinkRepo,
		SettingRepo: settingRepo,
		TagRepo:     tagRepo,

		AuditService:        auditService,
		ScopeService:        scopeService,
		SettingEventService: settingEventService,
		SettingInitService:  settingInitService,
		SettingService:      settingService,
		UserService:         userService,
		PermissionManager:   permissionManager,
	}
}
