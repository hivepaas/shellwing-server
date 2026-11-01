package userserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

func (s *service) DeleteUser(ctx context.Context, db database.IDB, user *entity.User) error {
	// Revoke target user's JWT, user can't access with the old token
	err := s.userTokenRepo.DelAll(ctx, user.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Delete ref resources in DB
	userIDs := []string{user.ID}

	// ACL permissions related to the user
	err = s.permissionManager.DeleteACLPermissionsByObjects(ctx, db, userIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// User files
	err = s.fileRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeUser, userIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Resource links
	err = s.resLinkRepo.DeleteAllByScope(ctx, db, base.ObjectScopeUser, userIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.DeleteAllByObjects(ctx, db, base.ObjectScopeUser, userIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Tasks
	err = s.taskRepo.DeleteAllByUsers(ctx, db, userIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// User photo
	if user.Photo != "" {
		err = s.binObjectRepo.DeleteByIDs(ctx, db, []string{user.Photo})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	user.DeletedAt = time.Now()
	err = s.userRepo.Update(ctx, db, user, bunex.UpdateColumns("deleted_at"))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
