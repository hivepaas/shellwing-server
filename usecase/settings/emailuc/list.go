package emailuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
)

func (uc *UC) ListEmail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.ListEmailReq,
) (*emaildto.ListEmailResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := emaildto.TransformEmails(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &emaildto.ListEmailResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
