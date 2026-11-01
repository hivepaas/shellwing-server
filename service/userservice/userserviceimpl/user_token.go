package userserviceimpl

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity/appentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

const (
	MFATokenExp           = 60 * time.Second
	MFATotpSetupTokenExp  = 180 * time.Second
	UserInviteTokenExp    = 7 * timeutil.Day
	PasswordResetTokenExp = 7 * timeutil.Day
	// SSOSignupTokenExp is how long somebody has to confirm the account a provider's sign-in offered them.
	SSOSignupTokenExp = 15 * time.Minute

	ssoSignupTokenKind = "sso-signup"
)

// GenerateMFAToken builds MFA token for using in the next step
func (s *service) GenerateMFAToken(userID string, mfaType base.MFAType, trustedDeviceID string) (string, error) {
	mfaToken, err := jwtsession.GenerateToken(&appentity.MFATokenClaims{
		Kind:            "mfa",
		UserID:          userID,
		MFAType:         mfaType,
		TrustedDeviceID: trustedDeviceID,
	}, MFATokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return mfaToken, nil
}

func (s *service) ParseMFAToken(token string) (*appentity.MFATokenClaims, error) {
	tokenClaims := &appentity.MFATokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "mfa" {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateMFATotpSetupToken builds MFA TOTP token for setting up TOTP secret
func (s *service) GenerateMFATotpSetupToken(userID string, toptSecret string) (string, error) {
	mfaToken, err := jwtsession.GenerateToken(&appentity.MFATotpSetupTokenClaims{
		Kind:   "mfa-setup",
		UserID: userID,
		Secret: toptSecret,
	}, MFATotpSetupTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return mfaToken, nil
}

func (s *service) ParseMFATotpSetupToken(token string) (*appentity.MFATotpSetupTokenClaims, error) {
	tokenClaims := &appentity.MFATotpSetupTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "mfa-setup" {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateUserInviteToken builds token for inviting users
func (s *service) GenerateUserInviteToken(userID string) (string, error) {
	token, err := jwtsession.GenerateToken(&appentity.UserInviteTokenClaims{
		Kind:   "user-invite",
		UserID: userID,
	}, UserInviteTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParseUserInviteToken(token string) (*appentity.UserInviteTokenClaims, error) {
	tokenClaims := &appentity.UserInviteTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "user-invite" {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GeneratePasswordResetToken builds token for resetting passwords
func (s *service) GeneratePasswordResetToken(userID string) (string, error) {
	token, err := jwtsession.GenerateToken(&appentity.PasswordResetTokenClaims{
		Kind:   "pwd-reset",
		UserID: userID,
	}, PasswordResetTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParsePasswordResetToken(token string) (*appentity.PasswordResetTokenClaims, error) {
	tokenClaims := &appentity.PasswordResetTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "pwd-reset" {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateSSOSignupToken signs what a provider's sign-in learned of somebody with no account, for the sign-up page.
func (s *service) GenerateSSOSignupToken(claims *appentity.SSOSignupTokenClaims) (string, error) {
	claims.Kind = ssoSignupTokenKind
	token, err := jwtsession.GenerateToken(claims, SSOSignupTokenExp)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParseSSOSignupToken(token string) (*appentity.SSOSignupTokenClaims, error) {
	tokenClaims := &appentity.SSOSignupTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != ssoSignupTokenKind || tokenClaims.Email == "" {
		return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}
