package keyauthuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
)

func (uc *UC) UpdateKeyAuth(
	ctx context.Context,
	auth *basedto.Auth,
	req *keyauthdto.UpdateKeyAuthReq,
) (*keyauthdto.UpdateKeyAuthResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	keyAuth := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: keyAuth.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			current, err := data.Setting.AsKeyAuth()
			if err != nil {
				return hperrors.Wrap(err)
			}
			req.KeepMaskedSecrets(keyAuth, current)

			err = pData.Setting.SetData(keyAuth)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &keyauthdto.UpdateKeyAuthResp{}, nil
}
