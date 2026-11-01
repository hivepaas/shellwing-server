package keyauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) GetKeyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.GetKeyAuthReq,
) (*keyauthdto.GetKeyAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := keyauthdto.TransformKeyAuth(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.GetKeyAuthResp{
		Data: respData,
	}, nil
}
