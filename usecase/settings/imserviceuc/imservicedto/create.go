package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

const (
	webhookURLMaxLen = 512
	tokenMaxLen      = 200
)

type CreateIMServiceReq struct {
	settings.CreateSettingReq
	*IMServiceBaseReq
}

type IMServiceBaseReq struct {
	Name    string             `json:"name"`
	Kind    base.IMServiceKind `json:"kind"`
	Slack   *IMSlackReq        `json:"slack"`
	Discord *IMDiscordReq      `json:"discord"`
}

func (req *IMServiceBaseReq) ToEntity() *entity.IMService {
	imService := &entity.IMService{}
	switch req.Kind {
	case base.IMServiceKindSlack:
		imService.Slack = req.Slack.ToEntity()
	case base.IMServiceKindDiscord:
		imService.Discord = req.Discord.ToEntity()
	}
	return imService
}

type IMSlackReq struct {
	Webhook string `json:"webhook"`
}

func (req *IMSlackReq) ToEntity() *entity.IMSlack {
	return &entity.IMSlack{
		Webhook: entity.NewEncryptedField(req.Webhook),
	}
}

func (req *IMSlackReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Webhook, true, 1, webhookURLMaxLen, field+"webhook")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Webhook, field+"webhook")...)
	return res
}

type IMDiscordReq struct {
	Webhook string `json:"webhook"`
}

func (req *IMDiscordReq) ToEntity() *entity.IMDiscord {
	return &entity.IMDiscord{
		Webhook: entity.NewEncryptedField(req.Webhook),
	}
}

func (req *IMDiscordReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Webhook, true, 1, webhookURLMaxLen, field+"webhook")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Webhook, field+"webhook")...)
	return res
}

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list. Only the selected kind is listed:
// ToEntity ignores the others, so a placeholder left in them is never stored.
func (req *IMServiceBaseReq) SecretFields() []basedto.SecretField {
	switch req.Kind {
	case base.IMServiceKindSlack:
		if req.Slack != nil {
			return []basedto.SecretField{{Path: "slack.webhook", Value: &req.Slack.Webhook}}
		}
	case base.IMServiceKindDiscord:
		if req.Discord != nil {
			return []basedto.SecretField{{Path: "discord.webhook", Value: &req.Discord.Webhook}}
		}
	}
	return nil
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *IMServiceBaseReq) KeepMaskedSecrets(imService, current *entity.IMService) {
	if current == nil {
		return
	}
	switch req.Kind {
	case base.IMServiceKindSlack:
		if req.Slack != nil && basedto.IsMaskedSecret(req.Slack.Webhook) &&
			imService.Slack != nil && current.Slack != nil {
			imService.Slack.Webhook = current.Slack.Webhook
		}
	case base.IMServiceKindDiscord:
		if req.Discord != nil && basedto.IsMaskedSecret(req.Discord.Webhook) &&
			imService.Discord != nil && current.Discord != nil {
			imService.Discord.Webhook = current.Discord.Webhook
		}
	}
}

func (req *IMServiceBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllIMServiceKinds, field+"kind")...)
	switch req.Kind {
	case base.IMServiceKindSlack:
		res = append(res, basedto.ValidateCond(req.Slack != nil, field+"slack")...)
		res = append(res, req.Slack.validate(field+"slack")...)
	case base.IMServiceKindDiscord:
		res = append(res, basedto.ValidateCond(req.Discord != nil, field+"discord")...)
		res = append(res, req.Discord.validate(field+"discord")...)
	}
	return res
}

func NewCreateIMServiceReq() *CreateIMServiceReq {
	return &CreateIMServiceReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateIMServiceResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
