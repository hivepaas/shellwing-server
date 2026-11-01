package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

var (
	mapProviderIcon = map[base.OAuthKind]string{
		base.OAuthKindGithubApp: "github", // icon mapping for special oauth types only
	}
)

func (uc *UC) GetLoginOptions(
	ctx context.Context,
	req *sessiondto.GetLoginOptionsReq,
) (*sessiondto.GetLoginOptionsResp, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeOAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectOrder("kind"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	cfg := config.Current()
	var resp []*sessiondto.LoginOptionResp
	for _, setting := range settings {
		oauthType := base.OAuthKind(setting.Kind)
		resp = append(resp, &sessiondto.LoginOptionResp{
			Type:    oauthType,
			Name:    setting.Name,
			Icon:    gofn.Coalesce(mapProviderIcon[oauthType], string(oauthType)),
			AuthURL: cfg.SsoAuthPath(setting.ID),
		})
	}

	return &sessiondto.GetLoginOptionsResp{
		Data: resp,
	}, nil
}
