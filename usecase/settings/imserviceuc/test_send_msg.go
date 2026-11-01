package imserviceuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/im/discord"
	"github.com/hivepaas/shellwing-server/pkg/im/slack"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
)

func (uc *UC) TestSendInstantMsg(
	ctx context.Context,
	auth *basedto.Auth,
	req *imservicedto.TestSendInstantMsgReq,
) (_ *imservicedto.TestSendInstantMsgResp, err error) {
	switch req.Kind {
	case base.IMServiceKindSlack:
		err = slack.NewClient().PostWebhook(ctx, req.Slack.Webhook, "", req.TestMsg)
	case base.IMServiceKindDiscord:
		_, err = discord.NewClient().WebhookExecute(ctx, req.Discord.Webhook, true, req.TestMsg)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imservicedto.TestSendInstantMsgResp{}, nil
}
