package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/service/settingeventservice"
)

type DeleteSettingReq struct {
	BaseSettingReq
	ID string `json:"-" mapstructure:"-"`
}

func (req *DeleteSettingReq) Validate() (validators []vld.Validator) {
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return
}

type DeleteSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type DeleteSettingData struct {
	BaseSettingData

	Setting       *entity.Setting
	ExtraLoadOpts []bunex.SelectQueryOption

	AfterLoading     func(context.Context, database.Tx, *DeleteSettingData) error
	BeforePersisting func(context.Context, database.Tx, *DeleteSettingData, *PersistingSettingDeletionData) error
	AfterPersisting  func(context.Context, database.Tx, *DeleteSettingData, *PersistingSettingDeletionData) error
}

type PersistingSettingDeletionData struct {
	Setting           *entity.Setting
	UpsertingSettings []*entity.Setting
}

func (uc *BaseUC) DeleteSetting(
	ctx context.Context,
	req *DeleteSettingReq,
	data *DeleteSettingData,
) (*DeleteSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadSettingForDeletion(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return hperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingDeletionData{}
		uc.prepareSettingDeletion(req, data, persistingData)
		if data.BeforePersisting != nil {
			if err := data.BeforePersisting(ctx, db, data, persistingData); err != nil {
				return hperrors.Wrap(err)
			}
		}

		err = uc.persistSettingDeletion(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.AfterPersisting != nil {
			if err := data.AfterPersisting(ctx, db, data, persistingData); err != nil {
				return hperrors.Wrap(err)
			}
		}

		// persistingData.Setting can be nil if the setting is imported, so fall
		// back to what was loaded - the entry must name what was deleted.
		err = uc.recordSettingAudit(ctx, db, &req.BaseSettingReq,
			base.AuditLogTypeSettingDelete, base.AuditLogSourceAPIDelete,
			gofn.Coalesce(persistingData.Setting, data.Setting), nil)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Fire delete event
		err = uc.SettingEventService.OnDelete(ctx, db, &settingeventservice.DeleteEvent{
			// persistingData.Setting can be nil if the setting is imported
			Setting: gofn.Coalesce(persistingData.Setting, data.Setting),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &DeleteSettingResp{}, nil
}

func (uc *BaseUC) loadSettingForDeletion(
	ctx context.Context,
	db database.IDB,
	req *DeleteSettingReq,
	data *DeleteSettingData,
) (err error) {
	if err = uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return hperrors.Wrap(err)
	}

	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectFor("UPDATE OF setting"),
	}
	loadOpts = append(loadOpts, data.ExtraLoadOpts...)

	setting, err := uc.loadSettingByID(ctx, db, &req.BaseSettingReq, req.ID,
		false, loadOpts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Setting = setting

	return nil
}

func (uc *BaseUC) prepareSettingDeletion(
	_ *DeleteSettingReq,
	data *DeleteSettingData,
	persistingData *PersistingSettingDeletionData,
) {
	timeNow := timeutil.NowUTC()
	data.Setting.UpdateVer++
	data.Setting.DeletedAt = timeNow
	persistingData.Setting = data.Setting
}

func (uc *BaseUC) persistSettingDeletion(
	ctx context.Context,
	db database.IDB,
	persistingData *PersistingSettingDeletionData,
) (err error) {
	err = uc.SettingRepo.Update(ctx, db, persistingData.Setting,
		bunex.UpdateColumns("deleted_at"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.SettingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
