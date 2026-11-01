package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/copier"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type GetNotificationReq struct {
	settings.GetSettingReq
}

func NewGetNotificationReq() *GetNotificationReq {
	return &GetNotificationReq{}
}

func (req *GetNotificationReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type NotificationResp struct {
	*settings.BaseSettingResp
	ViaEmail        *NotificationViaEmailResp   `json:"viaEmail,omitempty"`
	ViaSlack        *NotificationViaSlackResp   `json:"viaSlack,omitempty"`
	ViaDiscord      *NotificationViaDiscordResp `json:"viaDiscord,omitempty"`
	MinSendInterval timeutil.Duration           `json:"minSendInterval"`
}

type NotificationViaEmailResp struct {
	Enabled          bool                      `json:"enabled"`
	UseDefault       bool                      `json:"useDefault"`
	Sender           *settings.BaseSettingResp `json:"sender,omitempty"`
	ToProjectMembers bool                      `json:"toProjectMembers,omitempty"`
	ToProjectOwners  bool                      `json:"toProjectOwners,omitempty"`
	ToAllAdmins      bool                      `json:"toAllAdmins,omitempty"`
	ToAddresses      []string                  `json:"toAddresses,omitempty"`
}

type NotificationViaSlackResp struct {
	Enabled    bool                      `json:"enabled"`
	UseDefault bool                      `json:"useDefault"`
	Webhook    *settings.BaseSettingResp `json:"webhook,omitempty"`
}

type NotificationViaDiscordResp struct {
	Enabled    bool                      `json:"enabled"`
	UseDefault bool                      `json:"useDefault"`
	Webhook    *settings.BaseSettingResp `json:"webhook,omitempty"`
}

type GetNotificationResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *NotificationResp `json:"data"`
}

func TransformNotification(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *NotificationResp, err error) {
	config := setting.MustAsNotification()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if refObjects == nil {
		refObjects = &entity.RefObjects{}
	}

	viaEmail := resp.ViaEmail
	if viaEmail != nil && viaEmail.Sender != nil && viaEmail.Sender.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaEmail.Sender.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaEmail.Sender.ID, base.SettingTypeEmail)
		}
		viaEmail.Sender = itemResp
	}

	viaSlack := resp.ViaSlack
	if viaSlack != nil && viaSlack.Webhook != nil && viaSlack.Webhook.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaSlack.Webhook.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaSlack.Webhook.ID, base.SettingTypeIMService)
		}
		viaSlack.Webhook = itemResp
	}

	viaDiscord := resp.ViaDiscord
	if viaDiscord != nil && viaDiscord.Webhook != nil && viaDiscord.Webhook.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaDiscord.Webhook.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaDiscord.Webhook.ID, base.SettingTypeIMService)
		}
		viaDiscord.Webhook = itemResp
	}

	return resp, nil
}
