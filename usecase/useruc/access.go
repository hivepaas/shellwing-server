package useruc

import (
	"context"
	"slices"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

// authorizeAccessChanges checks the grants built into persistingData against the
// acting user's own permissions, and fills in the rows they replace.
//
// resourceTypes must cover every kind of grant persistingData carries. A kind
// left out of it is not merely unchecked here - see accessResourceTypesToReplace.
//
// The replaced set is deliberately not "everything the target has": a non-admin
// must not wipe grants they cannot see, so only rows they could revoke are
// cleared and the rest survive the update untouched.
func (uc *UC) authorizeAccessChanges(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	target *entity.User,
	resourceTypes []base.ResourceType,
	persistingData *userservice.PersistingUserData,
) error {
	if len(resourceTypes) == 0 {
		return nil
	}

	currentAccesses := make([]*entity.ACLPermission, 0, len(target.Accesses))
	for _, access := range target.Accesses {
		if slices.Contains(resourceTypes, access.ResourceType) {
			currentAccesses = append(currentAccesses, access)
		}
	}

	replaceable, err := uc.permissionManager.AuthorizeAccessChanges(ctx, db, auth,
		persistingData.UpsertingAccesses, currentAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, access := range replaceable {
		persistingData.DeletingAccesses = append(persistingData.DeletingAccesses,
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    target.ID,
				ResourceType: access.ResourceType,
				ResourceID:   access.ResourceID,
			})
	}
	return nil
}

// accessResourceTypesToReplace lists the resource types an operation replaces,
// based on which access lists the request carries.
//
// Every list the request can carry has to appear here, and a missing one fails in
// both directions at once: the rows it would replace are never cleared, so a
// grant can be handed out but never taken back, and - because an empty result
// makes authorizeAccessChanges return before it consults anyone - a request that
// carries only that list is written with no authorization check at all.
func accessResourceTypesToReplace(replaceModules, replaceCapabilities bool) []base.ResourceType {
	var resourceTypes []base.ResourceType
	if replaceModules {
		resourceTypes = append(resourceTypes, base.ResourceTypeModule)
	}
	if replaceCapabilities {
		resourceTypes = append(resourceTypes, base.ResourceTypeCapability)
	}
	return resourceTypes
}
