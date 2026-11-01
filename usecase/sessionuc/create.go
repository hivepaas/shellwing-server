package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

const (
	uidLen = 16
)

func (uc *UC) createSession(
	ctx context.Context,
	req *sessiondto.BaseCreateSessionReq,
) (resp *sessiondto.BaseCreateSessionResp, err error) {
	// Refused rather than recorded under an empty method. Every way into shellwing
	// ends here, so this is the one place that can make a login path added later
	// say how it authenticated - and a path that never says goes no further.
	if req.Method == "" {
		return nil, hperrors.NewArgumentInvalid("session login method")
	}
	// And the one place an account that may not sign in is stopped, whichever
	// way it came - a refresh included.
	if err = signInRefusal(req.User); err != nil {
		return nil, err
	}

	// If user is demo user, restrict permissions to READ only
	if req.User.IsDemoUser() {
		if req.AccessAction == nil {
			req.AccessAction = &base.AccessActions{}
		}
		req.AccessAction.Reset(true, false, false, false)
	}

	authClaims := &jwtsession.AuthClaims{
		UID:          gofn.RandTokenAsHex(uidLen),
		UserID:       req.User.ID,
		AccessAction: req.AccessAction,
	}

	// Decided per session rather than once at startup, because the answer depends
	// on who is signing in - and on a refresh it is re-decided from the account as
	// it is now, so an admin demoted mid-session lengthens on their next renewal
	// and a member promoted to admin shortens on theirs.
	exp := jwtsession.TokenExp(privilegedSession(req))

	// The deadline is measured from the login that began the session, so a
	// renewal cannot move it. Held to here rather than only checked, so the
	// tokens this hands out cannot outlive it either.
	startedAt := req.StartedAt
	if startedAt.IsZero() {
		startedAt = timeutil.NowUTC()
	}
	exp, alive := exp.Within(startedAt, timeutil.NowUTC())
	if !alive {
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTExpired).
			WithMsgLog("session reached its %v deadline and has to be signed into again", exp.Max)
	}
	authClaims.StartedAt = startedAt.Unix()

	resp = &sessiondto.BaseCreateSessionResp{}
	resp.AccessToken, err = jwtsession.GenerateAccessTokenWithExp(authClaims, exp.Access)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create access token")
	}
	resp.AccessTokenExp = authClaims.ExpiresAt.Time

	resp.RefreshToken, err = jwtsession.GenerateRefreshTokenWithExp(authClaims, exp.Refresh)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create refresh token")
	}
	resp.RefreshTokenExp = authClaims.ExpiresAt.Time

	// Stores the uid in cache, so we can revoke the token later
	err = uc.userTokenRepo.Set(ctx, authClaims.UserID, authClaims.UID, resp.RefreshTokenExp.Sub(timeutil.NowUTC()))
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to store token in cache")
	}

	// Last, and the login does not happen if it cannot be recorded. Nothing has
	// left this function yet, so refusing here costs the caller another sign-in;
	// the alternative is a live session nobody can account for, which is the one
	// thing this record exists to prevent.
	if err = uc.recordLogin(ctx, req.User, req.Method); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

// privilegedSession reports whether this session gets the admin lifetimes.
//
// An admin account. The role is the whole of the reason: an admin session can
// change anything the install has, so how long it stays usable unattended is
// worth deciding separately from everybody else's.
func privilegedSession(req *sessiondto.BaseCreateSessionReq) bool {
	return req.User != nil && req.User.IsAdmin()
}

// signInRefusal says why an account may not have a session: its access has
// expired, or it is not active - disabled, or still pending. Nil when it may.
func signInRefusal(user *entity.User) error {
	if user.IsAccessExpired() {
		return hperrors.Wrap(hperrors.ErrUserUnavailable).
			WithMsgLog("user access expired at: %v", user.AccessExpireAt)
	}
	if user.Status != base.UserStatusActive {
		return hperrors.Wrap(hperrors.ErrUserUnavailable).
			WithMsgLog("user status: %s", user.Status)
	}
	return nil
}
