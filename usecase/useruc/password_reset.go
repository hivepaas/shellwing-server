package useruc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) ResetPassword(
	ctx context.Context,
	req *userdto.ResetPasswordReq,
) (*userdto.ResetPasswordResp, error) {
	tokenClaims, err := uc.userService.ParsePasswordResetToken(req.Token)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, tokenClaims.UserID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if user.IsDemoUser() {
			return hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
		}

		err = uc.userService.ChangePassword(user, req.Password, userservice.SkipCheckingCurrentPassword)
		if err != nil {
			return hperrors.Wrap(err).WithMsgLog("failed to change password")
		}

		user.UpdatedAt = timeutil.NowUTC()
		err = uc.userRepo.Update(ctx, db, user,
			bunex.UpdateColumns("updated_at", "password"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// A password changed without a session, so the section is what tells this
		// apart from the account changing its own - and a reset nobody asked for
		// is the thing worth being able to see.
		return uc.recordUserChange(ctx, db, authOfUser(user), base.AuditLogTypeUserUpdate,
			auditSectionPasswordReset, user, nil)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.ResetPasswordResp{}, nil
}
