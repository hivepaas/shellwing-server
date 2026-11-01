package permissionimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
)

func (p *manager) checkModuleAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.ModuleAccessCheck,
) (bool, error) {
	return p.checkFlatResourceAccess(ctx, db, &check.BaseAccessCheck,
		base.ResourceTypeModule, string(check.Module))
}

// checkFlatResourceAccess answers a check against a single resource row, for the
// resource types that have no scope to walk up: the subject either holds the row
// or does not.
func (p *manager) checkFlatResourceAccess(
	ctx context.Context,
	db database.IDB,
	check *permission.BaseAccessCheck,
	resourceType base.ResourceType,
	resourceID string,
) (bool, error) {
	resource := &base.PermissionResource{
		SubjectType:  check.SubjectType,
		SubjectID:    check.SubjectID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	perms, err := p.aclPermissionRepo.ListByResources(ctx, db, []*base.PermissionResource{resource})
	if err != nil || len(perms) == 0 {
		return false, hperrors.Wrap(err)
	}

	// An empty resourceID matches nothing on purpose. It means the caller left the
	// module or capability unset, and answering that with the first row that has
	// an empty id would grant on a mistake.
	if resourceID == "" {
		return false, nil
	}
	for _, perm := range perms {
		if perm.ResourceType == resourceType && perm.ResourceID == resourceID {
			return p.hasPermission(perm, check), nil
		}
	}

	return false, nil
}
