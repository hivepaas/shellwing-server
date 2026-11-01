package imserviceuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) DeleteIMService(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.DeleteIMServiceReq,
) (*imservicedto.DeleteIMServiceResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.DeleteIMServiceResp{}, nil
}
