package notificationuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) DeleteNotification(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.DeleteNotificationReq,
) (*notificationdto.DeleteNotificationResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &notificationdto.DeleteNotificationResp{}, nil
}
