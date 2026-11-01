package useruc

import (
	"context"
	"errors"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) GetUserInviteInfo(
	ctx context.Context,
	_ *basedto.Auth,
	_ *userdto.GetUserInviteInfoReq,
) (*userdto.GetUserInviteInfoResp, error) {
	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.GetUserInviteInfoResp{
		Data: &userdto.UserInviteInfoResp{
			CanSendInviteEmails: emailSetting != nil && emailSetting.IsActive(),
		},
	}, nil
}
