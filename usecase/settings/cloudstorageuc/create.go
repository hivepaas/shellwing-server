package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) CreateCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.CreateCloudStorageReq,
) (*cloudstoragedto.CreateCloudStorageResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	cloudStorage := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: cloudStorage.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
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

	return &cloudstoragedto.CreateCloudStorageResp{
		Data: resp.Data,
	}, nil
}
