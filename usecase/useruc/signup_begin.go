package useruc

import (
	"context"
	"encoding/base64"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/totp"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) BeginUserSignup(
	ctx context.Context,
	req *userdto.BeginUserSignupReq,
) (*userdto.BeginUserSignupResp, error) {
	inviteToken, err := uc.userService.ParseUserInviteToken(req.InviteToken)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, inviteToken.UserID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if user.Status != base.UserStatusPending {
		return nil, hperrors.Wrap(hperrors.ErrUserStatusNotAllowAction).
			WithMsgLog("user '%s' is not required to signup", user.Email)
	}

	resp := &userdto.BeginUserSignupDataResp{
		Username:       user.Username,
		Email:          user.Email,
		FullName:       user.FullName,
		Role:           user.Role,
		SecurityOption: user.SecurityOption,
	}
	if !user.AccessExpireAt.IsZero() {
		resp.AccessExpireAt = &user.AccessExpireAt
	}

	// Generate TOTP secret and QR code for user to setup 2FA
	if user.SecurityOption == base.UserSecurityPassword2FA {
		secret, qrCode, err := totp.GenerateSecretAndQRCode(qrCodeImageSize)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp.MFATotpSecret = secret
		resp.QRCode = &userdto.MFATotpQRCodeResp{
			DataBase64: base64.StdEncoding.EncodeToString(qrCode.Bytes()),
			ImageType:  qrCodeImageType,
			ImageSize:  qrCodeImageSize,
		}
	}

	return &userdto.BeginUserSignupResp{
		Data: resp,
	}, nil
}
