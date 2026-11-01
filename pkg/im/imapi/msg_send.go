package imapi

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/im/discord"
	"github.com/hivepaas/shellwing-server/pkg/im/slack"
)

//nolint:gocognit
func SendMessage(
	ctx context.Context,
	setting *entity.Setting,
	msg string,
) error {
	if msg == "" {
		return nil
	}
	if setting == nil {
		return hperrors.NewMissing("IM service setting")
	}
	if setting.Type != base.SettingTypeIMService {
		return hperrors.Wrap(hperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}

	imService, err := setting.AsIMService()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if imService == nil {
		return hperrors.NewMissing("IM service setting")
	}

	switch base.IMServiceKind(setting.Kind) {
	case base.IMServiceKindSlack:
		if imService.Slack == nil {
			return hperrors.NewMissing("Slack setting")
		}
		webhookURL, err := imService.Slack.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = slack.NewClient().PostWebhook(ctx, webhookURL, "", msg)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.IMServiceKindDiscord:
		if imService.Discord == nil {
			return hperrors.NewMissing("Discord setting")
		}
		webhookURL, err := imService.Discord.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
		_, err = discord.NewClient().WebhookExecute(ctx, webhookURL, true, msg)
		if err != nil {
			return hperrors.Wrap(err)
		}

	default:
		return hperrors.Wrap(hperrors.ErrIMServiceUnsupported).WithParam("Name", setting.Kind)
	}

	return nil
}

func SendMessageWithRetry(
	ctx context.Context,
	setting *entity.Setting,
	msg string,
	retryMax int,
	retryDelay time.Duration,
) error {
	if msg == "" {
		return nil
	}
	if setting == nil {
		return hperrors.NewMissing("IM service setting")
	}
	return retryExecute(ctx, func() error {
		return SendMessage(ctx, setting, msg)
	}, base.IMServiceKind(setting.Kind), retryMax, retryDelay)
}
