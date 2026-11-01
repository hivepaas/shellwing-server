package useruc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/emailservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
	settingRepo   repository.SettingRepo
	userRepo      repository.UserRepo
	userTokenRepo cacherepository.UserTokenRepo
	rateLimitRepo cacherepository.RateLimitRepo

	auditService auditservice.Service
	emailService emailservice.Service
	userService  userservice.Service

	permissionManager permission.Manager
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
	settingRepo repository.SettingRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,
	rateLimitRepo cacherepository.RateLimitRepo,

	auditService auditservice.Service,
	emailService emailservice.Service,
	userService userservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
		settingRepo:   settingRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,
		rateLimitRepo: rateLimitRepo,

		auditService: auditService,
		emailService: emailService,
		userService:  userService,

		permissionManager: permissionManager,
	}
}
