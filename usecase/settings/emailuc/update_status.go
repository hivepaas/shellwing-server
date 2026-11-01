package emailuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
)

func (uc *UC) UpdateEmailStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.UpdateEmailStatusReq,
) (*emaildto.UpdateEmailStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &emaildto.UpdateEmailStatusResp{}, nil
}
