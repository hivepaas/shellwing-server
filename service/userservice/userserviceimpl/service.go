package userserviceimpl

import (
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/orgservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func New(
	binObjectRepo repository.BinObjectRepo,
	fileRepo repository.FileRepo,
	resLinkRepo repository.ResLinkRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	userRepo repository.UserRepo,
	userTokenRepo cacherepository.UserTokenRepo,

	permissionManager permission.Manager,
	orgService orgservice.Service,
) userservice.Service {
	return &service{
		binObjectRepo: binObjectRepo,
		fileRepo:      fileRepo,
		resLinkRepo:   resLinkRepo,
		settingRepo:   settingRepo,
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,

		permissionManager: permissionManager,
		orgService:        orgService,
	}
}

type service struct {
	binObjectRepo repository.BinObjectRepo
	fileRepo      repository.FileRepo
	resLinkRepo   repository.ResLinkRepo
	settingRepo   repository.SettingRepo
	taskRepo      repository.TaskRepo
	userRepo      repository.UserRepo
	userTokenRepo cacherepository.UserTokenRepo

	permissionManager permission.Manager
	orgService        orgservice.Service
}
