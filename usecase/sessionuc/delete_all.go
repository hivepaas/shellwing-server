package sessionuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

func (uc *UC) DeleteAllSessions(
	ctx context.Context,
	req *sessiondto.DeleteAllSessionsReq,
) (resp *sessiondto.DeleteAllSessionsResp, err error) {
	// Invalidate the old token to make it unusable
	err = uc.userTokenRepo.DelAll(ctx, req.User.AuthClaims.UserID)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to invalidate old token")
	}

	// Its apps too: a device would otherwise be given a new token at its next refresh.
	err = uc.deviceRepo.RevokeAllOfUser(ctx, uc.db, req.User.AuthClaims.UserID, timeutil.NowUTC())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// After the tokens are dead, so what is recorded is what happened. Worth more
	// than the single logout next door: this is the act that ends every other
	// session the account has, wherever they are.
	if err = uc.recordLogout(ctx, req.User, auditSectionLogoutAll); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sessiondto.DeleteAllSessionsResp{}, nil
}
