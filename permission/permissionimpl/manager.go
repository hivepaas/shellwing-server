package permissionimpl

import (
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

type manager struct {
	aclPermissionRepo repository.ACLPermissionRepo
	userRepo          repository.UserRepo

	auditService auditservice.Service
}

func NewManager(
	aclPermissionRepo repository.ACLPermissionRepo,
	userRepo repository.UserRepo,

	auditService auditservice.Service,
) permission.Manager {
	return &manager{
		aclPermissionRepo: aclPermissionRepo,
		userRepo:          userRepo,

		auditService: auditService,
	}
}
