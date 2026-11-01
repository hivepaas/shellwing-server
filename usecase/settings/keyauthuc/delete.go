package keyauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) DeleteKeyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.DeleteKeyAuthReq,
) (*keyauthdto.DeleteKeyAuthResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.DeleteKeyAuthResp{}, nil
}
