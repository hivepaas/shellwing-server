package emailuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
)

func (uc *UC) DeleteEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.DeleteEmailReq,
) (*emaildto.DeleteEmailResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &emaildto.DeleteEmailResp{}, nil
}
