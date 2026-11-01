package useruc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) UpdatePassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.UpdatePasswordReq,
) (*userdto.UpdatePasswordResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err := uc.userRepo.GetByID(ctx, db, auth.User.ID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		if user.SecurityOption == base.UserSecurityEnforceSSO {
			return hperrors.Wrap(hperrors.ErrActionNotAllowed).
				WithMsgLog("user authentication method is enforce-sso")
		}

		err = uc.userService.ChangePassword(user, req.NewPassword, req.CurrentPassword)
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

		return uc.recordUserChange(ctx, db, auth, base.AuditLogTypeUserUpdate,
			auditSectionPassword, user, nil)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.UpdatePasswordResp{}, nil
}
