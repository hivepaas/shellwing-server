package oauthuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc/oauthdto"
)

func (uc *UC) CreateOAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *oauthdto.CreateOAuthReq,
) (*oauthdto.CreateOAuthResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	oauth := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   gofn.Coalesce(req.Name, req.Organization),
		VerifyingRefIDs: oauth.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			err := pData.Setting.SetData(oauth)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &oauthdto.CreateOAuthResp{
		Data: &oauthdto.OAuthCreationResp{
			ID:          resp.Data.ID,
			CallbackURL: config.Current().SsoCallbackURL(resp.Data.ID),
		},
	}, nil
}
