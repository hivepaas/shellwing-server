package keyauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) ListKeyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.ListKeyAuthReq,
) (*keyauthdto.ListKeyAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := keyauthdto.TransformKeyAuths(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.ListKeyAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
