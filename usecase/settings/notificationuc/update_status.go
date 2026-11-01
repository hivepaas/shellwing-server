package notificationuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc/notificationdto"
)

func (uc *UC) UpdateNotificationStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *notificationdto.UpdateNotificationStatusReq,
) (*notificationdto.UpdateNotificationStatusResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &notificationdto.UpdateNotificationStatusResp{}, nil
}
