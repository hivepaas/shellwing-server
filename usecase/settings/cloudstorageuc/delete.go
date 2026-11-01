package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) DeleteCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.DeleteCloudStorageReq,
) (*cloudstoragedto.DeleteCloudStorageResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &cloudstoragedto.DeleteCloudStorageResp{}, nil
}
