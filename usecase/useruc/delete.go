package useruc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) DeleteUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.DeleteUserReq,
) (*userdto.DeleteUserResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		userData := &deleteUserData{}
		err := uc.loadUserDataForDelete(ctx, db, auth, req, userData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.userService.DeleteUser(ctx, db, userData.User)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// The role of the account that is going away, because after this there is
		// no row left to look it up in.
		return uc.recordUserChange(ctx, db, auth, base.AuditLogTypeUserDelete, "",
			userData.User, auditdetail.New().Set("role", userData.User.Role))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.DeleteUserResp{}, nil
}

type deleteUserData struct {
	User *entity.User
}

func (uc *UC) loadUserDataForDelete(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *userdto.DeleteUserReq,
	data *deleteUserData,
) error {
	user, err := uc.userRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.User = user

	if user.Role == base.UserRoleAdmin {
		if auth.User.Role != base.UserRoleAdmin {
			return hperrors.Wrap(hperrors.ErrActionNotAllowed).
				WithMsgLog("member user cannot delete admin user")
		}

		// Make sure there is at least one active admin user in the system after the deletion
		otherAdmins, _, err := uc.userRepo.List(ctx, db, nil,
			bunex.SelectWhere("id != ?", user.ID),
			bunex.SelectWhere("role = ?", base.UserRoleAdmin),
			bunex.SelectWhere("status = ?", base.UserStatusActive),
			bunex.SelectWhere("access_expire_at IS NULL OR access_expire_at > NOW()"),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if len(otherAdmins) == 0 {
			return hperrors.Wrap(hperrors.ErrActionNotAllowed).
				WithMsgLog("cannot delete the last admin user")
		}
	}

	return nil
}
