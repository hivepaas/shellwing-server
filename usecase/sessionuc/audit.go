package sessionuc

import (
	"context"
	"errors"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

// The ways in, as the entry's section. See AuditLogTypeUserLogin.
const (
	auditMethodPassword = "password"
	auditMethodPasscode = "passcode"
	auditMethodOAuth    = "oauth"
	auditMethodDevMode  = "dev-mode"
	auditMethodDevice   = "device"

	// auditMethodRefresh is a session renewing itself, and it is the one method
	// that records nothing - see recordLogin.
	auditMethodRefresh = "refresh"
)

// The two kinds of logout.
const (
	auditSectionLogoutCurrent = "current"
	auditSectionLogoutAll     = "all"
)

// Why a login was turned away.
//
// A short list of named reasons rather than the error text, because this is the
// column somebody counts: fifty wrong passwords against one account is a
// different picture from fifty unknown usernames from one address, and neither
// is visible if the reason is a sentence that varies.
const (
	auditReasonUnknownUser      = "unknown-user"
	auditReasonWrongPassword    = "wrong-password"
	auditReasonLockedOut        = "locked-out"
	auditReasonSSORequired      = "sso-required"
	auditReasonWrongPasscode    = "wrong-passcode"
	auditReasonTooManyPasscodes = "too-many-passcodes"
	auditReasonUserUnavailable  = "user-unavailable"
	// auditReasonEmailNotVerified is a sign-in with a provider that does not vouch for the email it gave.
	auditReasonEmailNotVerified = "email-not-verified"
)

// recordLogin records a session handed out to somebody who proved who they are.
//
// The user is the actor here, which is the one place in the system where that is
// established by this very request rather than carried in from a session.
func (uc *UC) recordLogin(ctx context.Context, user *entity.User, method string) error {
	// A refresh is not a login. Every session renews itself every few minutes for
	// as long as somebody keeps a tab open, and recording those would bury the
	// logins under their own renewals - the session was recorded when it was
	// created, and the renewal adds nothing a reader would ask for.
	if method == auditMethodRefresh {
		return nil
	}
	if user == nil {
		return hperrors.NewArgumentInvalid("audited login user")
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, uc.db, &auditservice.Entry{
		Type:     base.AuditLogTypeUserLogin,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.ID,
		Source:   base.AuditLogSourceAPICreate,
		Section:  method,
		Auth:     &basedto.Auth{User: &basedto.User{User: user}},
		ResType:  base.ResourceTypeUser,
		ResID:    user.ID,
		ResName:  user.Username,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// recordLoginDenied records an attempt that was refused.
//
// user is the account that was aimed at, and it is nil when the identifier
// offered matched none - which is also why nothing the caller typed is written
// down. A username box holds a password often enough that echoing it into an
// append-only table would be a way of collecting them, and "an identifier nobody
// has" plus the address it came from is the whole of what the entry has to say.
//
// No actor, on purpose: who was on the other end is exactly what this request
// failed to establish. The account, where there is one, is the entry's resource.
//
// Filed under the account's own scope so it shows on that user's history, and
// globally when there is no account to file it under.
func (uc *UC) recordLoginDenied(
	ctx context.Context,
	user *entity.User,
	method string,
	reason string,
) error {
	entry := &auditservice.Entry{
		Type:    base.AuditLogTypeUserLogin,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPICreate,
		Section: method,
		Result:  base.AuditLogResultDenied,
		ResType: base.ResourceTypeUser,
		Detail:  auditdetail.New().Set("reason", reason).String(),
	}
	if user != nil {
		entry.Scope = base.ObjectScopeUser
		entry.ObjectID = user.ID
		entry.ResID = user.ID
		entry.ResName = user.Username
	}

	if err := uc.auditService.Record(ctx, uc.db, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// recordLogout records a session being given up.
//
// The id comes from the session's own claims rather than from the loaded user,
// because the claims are the one part a logout is guaranteed to have: it is
// answering for the token in hand, not for a row it went and read.
func (uc *UC) recordLogout(ctx context.Context, user *basedto.User, section string) error {
	if user == nil || user.AuthClaims == nil {
		return hperrors.NewArgumentInvalid("audited logout user")
	}

	var username string
	if entityUser := user.Entity(); entityUser != nil {
		username = entityUser.Username
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, uc.db, &auditservice.Entry{
		Type:     base.AuditLogTypeUserLogout,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.AuthClaims.UserID,
		Source:   base.AuditLogSourceAPIDelete,
		Section:  section,
		Auth:     &basedto.Auth{User: user},
		ResType:  base.ResourceTypeUser,
		ResID:    user.AuthClaims.UserID,
		ResName:  username,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// loginRefusalReason names why a password login was turned down, and answers
// empty for anything that is not a refusal.
//
// The distinction is the point: redis being unreachable is the install failing,
// not somebody being turned away, and recording it among the attempts would put
// the system's own bad days into the column where break-in attempts are counted.
func loginRefusalReason(err error) string {
	switch {
	case errors.Is(err, hperrors.ErrSSORequired):
		return auditReasonSSORequired
	case errors.Is(err, hperrors.ErrTooManyLoginFailures):
		return auditReasonLockedOut
	case errors.Is(err, hperrors.ErrPasswordMismatched):
		return auditReasonWrongPassword
	case errors.Is(err, hperrors.ErrUserUnavailable):
		return auditReasonUserUnavailable
	}
	return ""
}

// refusePasscodeLogin records a second factor that was turned away and returns
// the error the caller sends back.
//
// The account is known here - the passcode step is reached only by getting past
// the password - so unlike the password refusals, this one can say whose
// second factor was missed.
func (uc *UC) refusePasscodeLogin(
	ctx context.Context,
	user *entity.User,
	reason string,
	err error,
) error {
	if e := uc.recordLoginDenied(ctx, user, auditMethodPasscode, reason); e != nil {
		return hperrors.Wrap(e)
	}
	return hperrors.Wrap(err)
}
