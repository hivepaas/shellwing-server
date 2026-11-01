package imserviceuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) UpdateIMServiceStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.UpdateIMServiceStatusReq,
) (*imservicedto.UpdateIMServiceStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.UpdateIMServiceStatusResp{}, nil
}
