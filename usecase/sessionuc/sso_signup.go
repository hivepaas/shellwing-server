package sessionuc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/markbates/goth"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/entity/appentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

const (
	// auditSectionSSOSignup is an account made by somebody who signed in with a provider and confirmed it.
	auditSectionSSOSignup = "sso-signup"
)

// ssoSignupOffer is what a sign-in with a provider answers somebody whose email no user has, where people may make
// their own accounts: a token for the sign-up page, which asks them to confirm before an account is made. Without
// the confirmation, a person who signed up with one email and signs in with a provider under another would find a
// second, empty account in its place.
func (uc *UC) ssoSignupOffer(oauthUser *goth.User, email string) (string, error) {
	if !config.Current().Users.SelfSignupEnabled {
		return "", nil
	}
	token, err := uc.userService.GenerateSSOSignupToken(&appentity.SSOSignupTokenClaims{
		Email:     email,
		FullName:  strings.TrimSpace(gofn.Coalesce(oauthUser.Name, oauthUser.NickName)),
		AvatarURL: gofn.If(len(oauthUser.AvatarURL) < externalAvatarURLMaxLen, oauthUser.AvatarURL, ""),
		Provider:  oauthUser.Provider,
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

// GetSSOSignupInfo is the account a sign-in with a provider offers to make, for the sign-up page to show.
func (uc *UC) GetSSOSignupInfo(
	_ context.Context,
	req *sessiondto.GetSSOSignupInfoReq,
) (*sessiondto.GetSSOSignupInfoResp, error) {
	claims, err := uc.ssoSignupClaims(req.Token)
	if err != nil {
		return nil, err
	}
	return &sessiondto.GetSSOSignupInfoResp{
		Data: &sessiondto.SSOSignupInfoDataResp{Email: claims.Email, FullName: claims.FullName},
	}, nil
}

// CompleteSSOSignup makes the account a sign-in with a provider offered, once its person confirmed it: an active
// member signing in with the provider, given the customer account that waits for the email, if one does. It
// answers with the account's session.
func (uc *UC) CompleteSSOSignup(
	ctx context.Context,
	req *sessiondto.CompleteSSOSignupReq,
) (*sessiondto.CompleteSSOSignupResp, error) {
	claims, err := uc.ssoSignupClaims(req.Token)
	if err != nil {
		return nil, err
	}

	var user *entity.User
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		user, err = uc.makeSSOSignupUser(ctx, db, claims)
		return err
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	session, err := uc.createSession(ctx, &sessiondto.BaseCreateSessionReq{User: user, Method: auditMethodOAuth})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &sessiondto.CompleteSSOSignupResp{Data: &sessiondto.CompleteSSOSignupDataResp{Session: session}}, nil
}

// makeSSOSignupUser makes the account, records it, and gives it the customer account that waits for its email.
func (uc *UC) makeSSOSignupUser(ctx context.Context, db database.Tx, claims *appentity.SSOSignupTokenClaims) (
	*entity.User, error) {
	existing, err := uc.userRepo.GetByEmail(ctx, db, claims.Email)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if existing != nil {
		// Made meanwhile - the same link confirmed twice, or a sign-up by email: it signs in as it is.
		return nil, hperrors.Wrap(hperrors.ErrEmailUnavailable).WithMsgLog("email '%s' has a user", claims.Email)
	}
	// db is the transaction that writes the user: see CheckUserSeat.
	if err = uc.licenseService.CheckUserSeat(ctx, db); err != nil {
		return nil, hperrors.Wrap(err)
	}
	username, err := userservice.UsernameForEmail(ctx, db, uc.userRepo, claims.Email)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	user := newSSOSignupUser(claims, username, timeutil.NowUTC())
	err = uc.userService.PersistUserData(ctx, db, &userservice.PersistingUserData{
		UpsertingUsers: []*entity.User{user},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	err = auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeUserCreate,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.ID,
		Source:   base.AuditLogSourceAPICreate,
		Section:  auditSectionSSOSignup,
		Auth:     authOfUser(user),
		ResType:  base.ResourceTypeUser,
		ResID:    user.ID,
		ResName:  user.Username,
		Detail:   auditdetail.New().Set("provider", claims.Provider).String(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return user, nil
}

// ssoSignupClaims are the sign-up token's, while people may make their own accounts.
func (uc *UC) ssoSignupClaims(token string) (*appentity.SSOSignupTokenClaims, error) {
	if !config.Current().Users.SelfSignupEnabled {
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowedByAdmin).WithMsgLog("self sign-up is turned off")
	}
	claims, err := uc.userService.ParseSSOSignupToken(token)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return claims, nil
}

// newSSOSignupUser is the account a provider's sign-in makes: a member like one who signed up by email, active at
// once - the provider vouched for the email - and with no password until they set one.
func newSSOSignupUser(claims *appentity.SSOSignupTokenClaims, username string, timeNow time.Time) *entity.User {
	return &entity.User{
		ID:             gofn.Must(ulid.NewStringULID()),
		Username:       username,
		Email:          claims.Email,
		FullName:       gofn.Coalesce(claims.FullName, username),
		Photo:          claims.AvatarURL,
		Role:           base.UserRoleMember,
		Status:         base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
		LastAccess:     timeNow,
	}
}

// authOfUser is the account acting for itself, where the provider's sign-in rather than a session is the proof.
func authOfUser(user *entity.User) *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: user}}
}
