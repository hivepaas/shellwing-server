package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) UpdateCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.UpdateCloudStorageReq,
) (*cloudstoragedto.UpdateCloudStorageResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	cloudStorage := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: cloudStorage.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			if err := uc.checkKeyAuth(ctx, db, cloudStorage); err != nil {
				return err
			}
			err := pData.Setting.SetData(cloudStorage)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &cloudstoragedto.UpdateCloudStorageResp{}, nil
}
