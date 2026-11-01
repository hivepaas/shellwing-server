package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/copier"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type GetIMServiceReq struct {
	settings.GetSettingReq
}

func NewGetIMServiceReq() *GetIMServiceReq {
	return &GetIMServiceReq{}
}

func (req *GetIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetIMServiceResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *IMServiceResp `json:"data"`
}

type IMServiceResp struct {
	*settings.BaseSettingResp
	Kind         base.IMServiceKind `json:"kind"`
	Slack        *IMSlackResp       `json:"slack,omitempty"`
	Discord      *IMDiscordResp     `json:"discord,omitempty"`
	SecretMasked bool               `json:"secretMasked,omitempty"`
}

type IMSlackResp struct {
	Webhook string `json:"webhook"`
}

func (resp *IMSlackResp) CopyWebhook(field entity.EncryptedField) error {
	resp.Webhook = field.String()
	return nil
}

type IMDiscordResp struct {
	Webhook string `json:"webhook"`
}

func (resp *IMDiscordResp) CopyWebhook(field entity.EncryptedField) error {
	resp.Webhook = field.String()
	return nil
}

func TransformIMService(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *IMServiceResp, err error) {
	config := setting.MustAsIMService()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Kind = base.IMServiceKind(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	switch {
	case config.Slack != nil:
		resp.SecretMasked = config.Slack.Webhook.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Slack.Webhook = basedto.MaskedSecret
		}
	case config.Discord != nil:
		resp.SecretMasked = config.Discord.Webhook.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Discord.Webhook = basedto.MaskedSecret
		}
	}

	return resp, nil
}
