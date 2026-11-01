package sessionuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/service/emailservice"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

func (uc *UC) LoginPasswordForgot(
	ctx context.Context,
	req *sessiondto.LoginPasswordForgotReq,
) (*sessiondto.LoginPasswordForgotResp, error) {
	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil {
		return nil, hperrors.NewNotFound("System email setting")
	}

	user, err := uc.userRepo.GetByUsernameOrEmail(ctx, uc.db, req.Email, req.Email,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowed).
			WithMsgLog("user authentication method is enforce-sso")
	}

	token, err := uc.userService.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to generate password reset token")
	}

	resetLink := config.Current().DashboardPasswordResetURL(user.ID, token)

	email, err := emailSetting.AsEmail()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.emailService.SendMailPasswordReset(ctx, uc.db, &emailservice.EmailDataPasswordReset{
		BaseTemplateData: emailservice.BaseTemplateData{
			Email:      email,
			Recipients: []string{user.Email},
			Subject:    "[shellwing] Password reset",
		},
		ResetPasswordLink: resetLink,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sessiondto.LoginPasswordForgotResp{}, nil
}
