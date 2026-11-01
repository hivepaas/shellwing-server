package sessionuc

import (
	"context"
	"testing"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/entity/appentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

// signupUsers signs sign-up tokens as their email, keeps the users it persists, and hands out passcode tokens.
type signupUsers struct {
	mfaTokens
	tokens    map[string]*appentity.SSOSignupTokenClaims
	persisted []*entity.User
}

func (s *signupUsers) GenerateSSOSignupToken(claims *appentity.SSOSignupTokenClaims) (string, error) {
	token := "signup:" + claims.Email
	s.tokens[token] = claims
	return token, nil
}

func (s *signupUsers) ParseSSOSignupToken(token string) (*appentity.SSOSignupTokenClaims, error) {
	if claims, found := s.tokens[token]; found {
		return claims, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrTokenInvalid)
}

func (s *signupUsers) PersistUserData(_ context.Context, _ database.IDB, data *userservice.PersistingUserData) error {
	s.persisted = append(s.persisted, data.UpsertingUsers...)
	return nil
}

func newSignupUCTest(t *testing.T, selfSignup bool) (*UC, *spyAuditService, *signupUsers) {
	t.Helper()
	uc, audit := newOAuthUCTest(t)
	config.Current().Users.SelfSignupEnabled = selfSignup
	users := &signupUsers{tokens: map[string]*appentity.SSOSignupTokenClaims{}}
	uc.userService = users
	uc.cacheMfaPasscodeRepo = &passcodeAttempts{}
	return uc, audit, users
}

func annaSignsInWithGoogle() *sessiondto.CreateOAuthSessionReq {
	return &sessiondto.CreateOAuthSessionReq{User: &goth.User{Provider: "set-google", Email: "Anna@example.com",
		Name: "Anna Smith", AvatarURL: "https://example.com/anna.png", RawData: map[string]any{"verified_email": true}}}
}

// Somebody with no account, where people may make their own, is offered one to confirm: nothing is made yet, and
// nobody was refused.
func TestCreateOAuthSessionOffersAnAccount(t *testing.T) {
	uc, audit, users := newSignupUCTest(t, true)

	resp, err := uc.CreateOAuthSession(context.Background(), annaSignsInWithGoogle())

	require.NoError(t, err)
	assert.Equal(t, "signup:anna@example.com", resp.SignupToken)
	assert.Empty(t, resp.AccessToken)
	assert.Empty(t, users.persisted)
	assert.Empty(t, audit.entries)
	claims := users.tokens[resp.SignupToken]
	assert.Equal(t, "Anna Smith", claims.FullName)
	assert.Equal(t, "set-google", claims.Provider)
}

// Where people may not make their own accounts, nobody is offered one: the sign-in is refused, and recorded so.
func TestCreateOAuthSessionOffersNoAccountWhenSignupIsOff(t *testing.T) {
	uc, audit, users := newSignupUCTest(t, false)

	_, err := uc.CreateOAuthSession(context.Background(), annaSignsInWithGoogle())

	require.ErrorIs(t, err, hperrors.ErrOAuthUserNotFound)
	assert.Empty(t, users.tokens)
	require.Len(t, audit.entries, 1)
	assert.Contains(t, audit.entries[0].Detail, auditReasonUnknownUser)
}

func TestGetSSOSignupInfo(t *testing.T) {
	uc, _, users := newSignupUCTest(t, true)
	users.tokens["signup:anna@example.com"] = &appentity.SSOSignupTokenClaims{Email: "anna@example.com",
		FullName: "Anna Smith"}
	req := sessiondto.NewGetSSOSignupInfoReq()
	req.Token = "signup:anna@example.com"

	resp, err := uc.GetSSOSignupInfo(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, &sessiondto.SSOSignupInfoDataResp{Email: "anna@example.com", FullName: "Anna Smith"}, resp.Data)

	req.Token = "signup:forged"
	_, err = uc.GetSSOSignupInfo(context.Background(), req)
	require.ErrorIs(t, err, hperrors.ErrTokenInvalid)

	config.Current().Users.SelfSignupEnabled = false
	req.Token = "signup:anna@example.com"
	_, err = uc.GetSSOSignupInfo(context.Background(), req)
	require.ErrorIs(t, err, hperrors.ErrActionNotAllowedByAdmin, "a token from before sign-up was turned off")
}

// The confirmed account is an active member with no password, recorded as made by a provider's sign-up.
func TestMakeSSOSignupUser(t *testing.T) {
	uc, audit, users := newSignupUCTest(t, true)
	claims := &appentity.SSOSignupTokenClaims{Email: "anna@example.com", FullName: "Anna Smith",
		AvatarURL: "https://example.com/anna.png", Provider: "set-google"}

	user, err := uc.makeSSOSignupUser(context.Background(), nil, claims)

	require.NoError(t, err)
	require.Equal(t, []*entity.User{user}, users.persisted)
	assert.Equal(t, "anna", user.Username)
	assert.Equal(t, "anna@example.com", user.Email)
	assert.Equal(t, "Anna Smith", user.FullName)
	assert.Equal(t, "https://example.com/anna.png", user.Photo)
	assert.Equal(t, base.UserRoleMember, user.Role)
	assert.Equal(t, base.UserStatusActive, user.Status)
	assert.Empty(t, user.Password)
	require.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogTypeUserCreate, audit.entries[0].Type)
	assert.Equal(t, auditSectionSSOSignup, audit.entries[0].Section)

	// An email that has a user by now - the link confirmed twice - makes nothing more.
	uc.userRepo.(*userByEmail).users["anna@example.com"] = user
	_, err = uc.makeSSOSignupUser(context.Background(), nil, claims)
	require.ErrorIs(t, err, hperrors.ErrEmailUnavailable)
	assert.Len(t, users.persisted, 1)
}
