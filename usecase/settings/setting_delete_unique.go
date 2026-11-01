package settings

import (
	"context"

	vld "github.com/tiendc/go-validator"

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

type DeleteUniqueSettingReq struct {
	BaseSettingReq
}

func (req *DeleteUniqueSettingReq) Validate() (validators []vld.Validator) {
	return
}

type DeleteUniqueSettingResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type DeleteUniqueSettingData struct {
	BaseSettingData

	Setting       *entity.Setting
	ExtraLoadOpts []bunex.SelectQueryOption

	AfterLoading     func(context.Context, database.Tx, *DeleteUniqueSettingData) error
	BeforePersisting func(context.Context, database.Tx, *DeleteUniqueSettingData, *PersistingSettingDeletionData) error
	AfterPersisting  func(context.Context, database.Tx, *DeleteUniqueSettingData, *PersistingSettingDeletionData) error
}

func (uc *BaseUC) DeleteUniqueSetting(
	ctx context.Context,
	req *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
) (*DeleteUniqueSettingResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		err := uc.loadUniqueSettingForDeletion(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if data.AfterLoading != nil {
			if err := data.AfterLoading(ctx, db, data); err != nil {
				return hperrors.Wrap(err)
			}
		}

		persistingData := &PersistingSettingDeletionData{}
		uc.prepareUniqueSettingDeletion(req, data, persistingData)
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

		err = uc.recordSettingAudit(ctx, db, &req.BaseSettingReq,
			base.AuditLogTypeSettingDelete, base.AuditLogSourceAPIDelete,
			persistingData.Setting, nil)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Fire delete event
		err = uc.SettingEventService.OnDelete(ctx, db, &settingeventservice.DeleteEvent{
			Setting: persistingData.Setting,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &DeleteUniqueSettingResp{}, nil
}

func (uc *BaseUC) loadUniqueSettingForDeletion(
	ctx context.Context,
	db database.IDB,
	req *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
) (err error) {
	if err = uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return hperrors.Wrap(err)
	}

	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectFor("UPDATE OF setting"),
	}
	loadOpts = append(loadOpts, data.ExtraLoadOpts...)

	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, req.Type, false, loadOpts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Setting = setting

	return nil
}

func (uc *BaseUC) prepareUniqueSettingDeletion(
	_ *DeleteUniqueSettingReq,
	data *DeleteUniqueSettingData,
	persistingData *PersistingSettingDeletionData,
) {
	timeNow := timeutil.NowUTC()
	data.Setting.UpdateVer++
	data.Setting.DeletedAt = timeNow
	persistingData.Setting = data.Setting
}
