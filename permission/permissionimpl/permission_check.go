package permissionimpl

import (
	"context"
	"maps"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
)

func (p *manager) CheckAccess(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	check permission.AccessCheck,
) (hasPerm bool, err error) {
	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return true, nil
	}

	check.InitSubject(auth)

	var allowedResources map[base.ResourceType][]string
	switch chk := check.(type) {
	case *permission.ModuleAccessCheck:
		hasPerm, err = p.checkModuleAccess(ctx, db, chk)
	case *permission.GeneralResourceAccessCheck:
		hasPerm, allowedResources, err = p.checkResourceAccess(ctx, db, chk)
	case *permission.CapabilityCheck:
		hasPerm, err = p.checkCapability(ctx, db, chk)
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}

	if len(allowedResources) > 0 {
		hasPerm = true
		if auth.AllowedResources == nil {
			auth.AllowedResources = allowedResources
		} else {
			auth.AllowedResources = make(map[base.ResourceType][]string, len(allowedResources))
			maps.Copy(auth.AllowedResources, allowedResources)
		}
	}

	return hasPerm, nil
}

func (p *manager) CheckAccessOnSetting(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	check permission.AccessCheck,
	setting *entity.Setting,
) (bool, error) {
	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return true, nil
	}

	switch setting.Scope {
	case base.ObjectScopeUser:
		return p.CheckAccess(ctx, db, auth, &permission.ModuleAccessCheck{
			BaseAccessCheck: *check.GetBase(),
			Module:          base.ResourceModuleUser,
		})
	case base.ObjectScopeGlobal:
		return p.CheckAccess(ctx, db, auth, &permission.ModuleAccessCheck{
			BaseAccessCheck: *check.GetBase(),
			Module:          base.ResourceModuleSettings,
		})
	default:
		return false, hperrors.Wrap(hperrors.ErrFileScopeUnsupported).WithParam("Scope", setting.Scope)
	}
}
