package oauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) ListOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.ListOAuthReq,
) (*oauthdto.ListOAuthResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &oauthdto.OAuthTransformInput{
		RefObjects:      resp.RefObjects,
		BaseCallbackURL: config.Current().SsoBaseCallbackURL(),
	}
	respData, err := oauthdto.TransformOAuths(resp.Data, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &oauthdto.ListOAuthResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
