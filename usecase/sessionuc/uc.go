package sessionuc

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
	cacheLoginAttemptRepo cacherepository.LoginAttemptRepo
	cacheMfaPasscodeRepo  cacherepository.MFAPasscodeRepo
	db                    *database.DB

	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo
	settingRepo            repository.SettingRepo
	systemStatusRepo       repository.SystemStatusRepo
	userRepo               repository.UserRepo
	userTokenRepo          cacherepository.UserTokenRepo

	auditService auditservice.Service
	emailService emailservice.Service
	userService  userservice.Service

	permissionManager permission.Manager
}

func New(
	cacheLoginAttemptRepo cacherepository.LoginAttemptRepo,
	cacheMfaPasscodeRepo cacherepository.MFAPasscodeRepo,
	db *database.DB,

	loginTrustedDeviceRepo repository.LoginTrustedDeviceRepo,
	settingRepo repository.SettingRepo,
	systemStatusRepo repository.SystemStatusRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	auditService auditservice.Service,
	emailService emailservice.Service,
	userService userservice.Service,

	permissionManager permission.Manager,
) *UC {
	return &UC{
		cacheLoginAttemptRepo: cacheLoginAttemptRepo,
		cacheMfaPasscodeRepo:  cacheMfaPasscodeRepo,
		db:                    db,

		loginTrustedDeviceRepo: loginTrustedDeviceRepo,
		settingRepo:            settingRepo,
		systemStatusRepo:       systemStatusRepo,
		userRepo:               userRepo,
		userTokenRepo:          userTokenRepo,

		auditService: auditService,
		emailService: emailService,
		userService:  userService,

		permissionManager: permissionManager,
	}
}
