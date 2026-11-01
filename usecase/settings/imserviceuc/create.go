package imserviceuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) CreateIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.CreateIMServiceReq,
) (*imservicedto.CreateIMServiceResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	imPlatform := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: imPlatform.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			err := pData.Setting.SetData(imPlatform)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.CreateIMServiceResp{
		Data: resp.Data,
	}, nil
}
