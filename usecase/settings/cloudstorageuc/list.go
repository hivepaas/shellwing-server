package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
)

func (uc *UC) ListCloudStorage(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.ListCloudStorageReq,
) (*cloudstoragedto.ListCloudStorageResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := cloudstoragedto.TransformCloudStorages(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &cloudstoragedto.ListCloudStorageResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
