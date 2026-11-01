package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) GetCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.GetCloudStorageReq,
) (*cloudstoragedto.GetCloudStorageResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := cloudstoragedto.TransformCloudStorage(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &cloudstoragedto.GetCloudStorageResp{
		Data: respData,
	}, nil
}
