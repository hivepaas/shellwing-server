package emailuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
)

func (uc *UC) GetEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.GetEmailReq,
) (*emaildto.GetEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := emaildto.TransformEmail(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &emaildto.GetEmailResp{
		Data: respData,
	}, nil
}
