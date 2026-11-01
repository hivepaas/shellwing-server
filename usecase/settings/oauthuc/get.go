package oauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) GetOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.GetOAuthReq,
) (*oauthdto.GetOAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &oauthdto.OAuthTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current().SsoBaseCallbackURL(),
	}
	respData, err := oauthdto.TransformOAuth(resp.Data, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &oauthdto.GetOAuthResp{
		Data: respData,
	}, nil
}
