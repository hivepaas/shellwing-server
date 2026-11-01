package settingserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/settingservice"
)

func (s *service) PersistSettingData(ctx context.Context, db database.IDB,
	persistingData *settingservice.PersistingSettingData) error {
	// Deletes data
	err := s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.DeletingAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Persists data
	// Settings
	err = s.settingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Accesses
	err = s.permissionManager.UpdateACLPermissions(ctx, db, persistingData.UpsertingAccesses)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
