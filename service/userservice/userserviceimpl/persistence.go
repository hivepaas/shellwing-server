package userserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func (s *service) PersistUserData(ctx context.Context, db database.IDB,
	persistingData *userservice.PersistingUserData) error {
	// Persists data
	// Users
	err := s.userRepo.UpsertMulti(ctx, db, persistingData.UpsertingUsers,
		entity.UserUpsertingConflictCols, entity.UserUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Binaries
	err = s.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Remove accesses
	err = s.permissionManager.DeleteACLPermissions(ctx, db, persistingData.DeletingAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Project/App/... accesses
	err = s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.UpsertingAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
