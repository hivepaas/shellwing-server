package useruc

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/reqinfo"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/emailservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

// The fakes embed the interfaces they stand in for, so a call the sign-up request is not expected to make
// panics on the nil interface instead of passing unnoticed.

type fakeSignupUserRepo struct {
	repository.UserRepo
	byEmail map[string]*entity.User
}

func (f *fakeSignupUserRepo) GetByEmail(
	_ context.Context, _ database.IDB, email string, _ ...bunex.SelectQueryOption,
) (*entity.User, error) {
	if user, ok := f.byEmail[email]; ok {
		return user, nil
	}
	return nil, hperrors.NewNotFound("User")
}

type fakeSignupUserService struct {
	userservice.Service
	persisted []*userservice.PersistingUserData
	tokensFor []string
}

func (f *fakeSignupUserService) GenerateUserInviteToken(userID string) (string, error) {
	f.tokensFor = append(f.tokensFor, userID)
	return "token-for-" + userID, nil
}

func (f *fakeSignupUserService) PersistUserData(
	_ context.Context, _ database.IDB, data *userservice.PersistingUserData,
) error {
	f.persisted = append(f.persisted, data)
	return nil
}

type fakeSignupEmailService struct {
	emailservice.Service
	noSystemEmail bool
	sendErr       error
	sent          []*emailservice.EmailDataSignupVerify
}

func (f *fakeSignupEmailService) GetDefaultSystemEmail(_ context.Context, _ database.IDB) (*entity.Setting, error) {
	if f.noSystemEmail {
		return nil, hperrors.NewNotFound("Email setting")
	}
	return &entity.Setting{Type: base.SettingTypeEmail, Data: `{"smtp":{"host":"smtp.example.com","port":587}}`}, nil
}

func (f *fakeSignupEmailService) SendMailSignupVerify(
	_ context.Context, _ database.IDB, data *emailservice.EmailDataSignupVerify,
) error {
	f.sent = append(f.sent, data)
	return f.sendErr
}

type fakeRateLimitRepo struct {
	counts map[string]int64
}

func (f *fakeRateLimitRepo) Hit(_ context.Context, key string, window time.Duration) (int64, error) {
	if window != signupRequestWindow {
		return 0, errors.New("unexpected window")
	}
	f.counts[key]++
	return f.counts[key], nil
}

type signupTest struct {
	uc     *UC
	users  *fakeSignupUserRepo
	svc    *fakeSignupUserService
	email  *fakeSignupEmailService
	limits *fakeRateLimitRepo
	audit  *fakeAuditService
}

func newSignupTest(t *testing.T, selfSignupEnabled bool, users ...*entity.User) *signupTest {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{
		WebBaseURL: "https://shellwing.example.com",
		Users:      config.Users{SelfSignupEnabled: selfSignupEnabled},
	})
	t.Cleanup(func() { config.SetCurrent(previous) })

	st := &signupTest{
		users:  &fakeSignupUserRepo{byEmail: map[string]*entity.User{}},
		svc:    &fakeSignupUserService{},
		email:  &fakeSignupEmailService{},
		limits: &fakeRateLimitRepo{counts: map[string]int64{}},
		audit:  &fakeAuditService{},
	}
	for _, user := range users {
		st.users.byEmail[user.Email] = user
	}
	st.uc = &UC{
		userRepo:      st.users,
		rateLimitRepo: st.limits,
		auditService:  st.audit,
		emailService:  st.email,
		userService:   st.svc,
	}
	return st
}

func signupCtx(clientIP string) context.Context {
	return reqinfo.NewContext(context.Background(), &reqinfo.RequestInfo{ClientIP: clientIP})
}

func signupReq(email string) *userdto.RequestUserSignupReq {
	return &userdto.RequestUserSignupReq{Email: email, FullName: "Some Person", Next: "/account/trial?plan=hive"}
}

func userWithStatus(id, email string, status base.UserStatus) *entity.User {
	return &entity.User{ID: id, Username: id, Email: email, Status: status, Role: base.UserRoleMember}
}

func TestSignupRequestActionFor(t *testing.T) {
	assert.Equal(t, signupRequestCreate, signupRequestActionFor(nil))
	assert.Equal(t, signupRequestResend, signupRequestActionFor(&entity.User{Status: base.UserStatusPending}))
	assert.Equal(t, signupRequestIgnore, signupRequestActionFor(&entity.User{Status: base.UserStatusActive}))
	assert.Equal(t, signupRequestIgnore, signupRequestActionFor(&entity.User{Status: base.UserStatusDisabled}))
	assert.Equal(t, signupRequestIgnore, signupRequestActionFor(&entity.User{Status: "future-status"}))
}

// Off, the form is refused before it counts or looks up anything.
func TestRequestUserSignupRefusedWhenDisabled(t *testing.T) {
	st := newSignupTest(t, false)

	resp, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq("new@example.com"))

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, hperrors.ErrActionNotAllowedByAdmin)
	assert.Empty(t, st.limits.counts)
	assert.Empty(t, st.email.sent)
}

// An unfinished account - a sign-up or an invitation nobody completed - gets a new link, carrying where the
// web goes afterwards.
func TestRequestUserSignupResendsToPendingAccount(t *testing.T) {
	pending := userWithStatus("user-1", "pending@example.com", base.UserStatusPending)
	st := newSignupTest(t, true, pending)

	resp, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq("pending@example.com"))

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, st.svc.persisted, "nothing is written for an account that exists")

	require.Len(t, st.email.sent, 1)
	mail := st.email.sent[0]
	assert.Equal(t, []string{"pending@example.com"}, mail.Recipients)
	require.NotNil(t, mail.Email)
	assert.Equal(t, "smtp.example.com", mail.Email.SMTP.Host)
	link, err := url.Parse(mail.SignupLink)
	require.NoError(t, err)
	assert.Equal(t, "shellwing.example.com", link.Host)
	assert.Equal(t, "/auth/sign-up", link.Path)
	assert.Equal(t, "token-for-user-1", link.Query().Get("token"))
	assert.Equal(t, "/account/trial?plan=hive", link.Query().Get("next"))

	// A link went out, asked for by nobody the request can name.
	require.Len(t, st.audit.entries, 1)
	entry := st.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeUserUpdate, entry.Type)
	assert.Equal(t, auditSectionSelfSignup, entry.Section)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Equal(t, "user-1", entry.ObjectID)
	assert.Nil(t, entry.Auth)
	assert.Contains(t, entry.Detail, `"linkResent":true`)
}

// An account in use is sent nothing, and the answer is the same as for a new address.
func TestRequestUserSignupSendsNothingToAccountInUse(t *testing.T) {
	for _, status := range []base.UserStatus{base.UserStatusActive, base.UserStatusDisabled} {
		t.Run(string(status), func(t *testing.T) {
			user := userWithStatus("user-1", "taken@example.com", status)
			st := newSignupTest(t, true, user)

			resp, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq("taken@example.com"))

			require.NoError(t, err)
			assert.Equal(t, &userdto.RequestUserSignupResp{}, resp)
			assert.Empty(t, st.email.sent)
			assert.Empty(t, st.svc.tokensFor, "no link is made")
			assert.Empty(t, st.svc.persisted)

			require.Len(t, st.audit.entries, 1)
			entry := st.audit.entries[0]
			assert.Equal(t, base.AuditLogResultDenied, entry.Result)
			assert.Equal(t, auditSectionSelfSignup, entry.Section)
			assert.Equal(t, "user-1", entry.ObjectID)
			assert.Contains(t, entry.Detail, auditReasonAccountExists)
			assert.NotContains(t, entry.Detail, "taken@example.com")
		})
	}
}

// Without mail set up, every address gets the same refusal: answering an account in use with success and a
// new one with an error would tell them apart.
func TestRequestUserSignupWithoutSystemEmail(t *testing.T) {
	for _, email := range []string{"taken@example.com", "pending@example.com", "new@example.com"} {
		st := newSignupTest(t, true,
			userWithStatus("user-1", "taken@example.com", base.UserStatusActive),
			userWithStatus("user-2", "pending@example.com", base.UserStatusPending))
		st.email.noSystemEmail = true

		_, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq(email))

		assert.ErrorIs(t, err, hperrors.ErrActionNotAllowedByAdmin, email)
		assert.Empty(t, st.svc.persisted, email)
		assert.Empty(t, st.audit.entries, email)
	}
}

func TestRequestUserSignupMailFailure(t *testing.T) {
	st := newSignupTest(t, true, userWithStatus("user-1", "pending@example.com", base.UserStatusPending))
	st.email.sendErr = errors.New("smtp down")

	_, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq("pending@example.com"))

	assert.Error(t, err)
}

// Five an hour from one address, whatever the emails; three an hour for one email, whatever the addresses.
func TestRequestUserSignupRateLimits(t *testing.T) {
	t.Run("per address", func(t *testing.T) {
		var users []*entity.User
		for i := range signupRequestsPerIP + 1 {
			users = append(users, userWithStatus("user-"+string(rune('a'+i)),
				string(rune('a'+i))+"@example.com", base.UserStatusActive))
		}
		st := newSignupTest(t, true, users...)
		ctx := signupCtx("203.0.113.7")

		for _, user := range users[:signupRequestsPerIP] {
			_, err := st.uc.RequestUserSignup(ctx, signupReq(user.Email))
			require.NoError(t, err, user.Email)
		}
		emailKeys := len(st.limits.counts) - 1

		_, err := st.uc.RequestUserSignup(ctx, signupReq(users[signupRequestsPerIP].Email))
		assert.ErrorIs(t, err, hperrors.ErrTooManyRequests)
		assert.Len(t, st.limits.counts, emailKeys+1, "a refused address does not use up the email's limit")

		// Another address is not held back.
		_, err = st.uc.RequestUserSignup(signupCtx("198.51.100.1"), signupReq(users[signupRequestsPerIP].Email))
		assert.NoError(t, err)
	})

	t.Run("per email", func(t *testing.T) {
		st := newSignupTest(t, true, userWithStatus("user-1", "taken@example.com", base.UserStatusActive))

		for i := range signupRequestsPerEmail {
			_, err := st.uc.RequestUserSignup(signupCtx("198.51.100."+string(rune('1'+i))),
				signupReq("taken@example.com"))
			require.NoError(t, err)
		}
		_, err := st.uc.RequestUserSignup(signupCtx("198.51.100.9"), signupReq("taken@example.com"))
		assert.ErrorIs(t, err, hperrors.ErrTooManyRequests)

		// The limit is kept without the address itself.
		for key := range st.limits.counts {
			assert.NotContains(t, key, "taken@example.com")
			assert.True(t, strings.HasPrefix(key, "signup-request:"), key)
		}
	})

	t.Run("checked before anything else", func(t *testing.T) {
		st := newSignupTest(t, true)
		st.limits.counts["signup-request:ip:203.0.113.7"] = signupRequestsPerIP

		_, err := st.uc.RequestUserSignup(signupCtx("203.0.113.7"), signupReq("new@example.com"))

		assert.ErrorIs(t, err, hperrors.ErrTooManyRequests)
		assert.Empty(t, st.audit.entries)
		assert.Empty(t, st.email.sent)
	})
}

// A self sign-up is a member with nothing granted and a password to choose, waiting for its link.
func TestNewSelfSignupUser(t *testing.T) {
	now := timeutil.NowUTC()
	user := newSelfSignupUser(&userdto.RequestUserSignupReq{Email: "new@example.com", FullName: "New Person"},
		"new", now)

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "new", user.Username)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "New Person", user.FullName)
	assert.Equal(t, base.UserRoleMember, user.Role)
	assert.Equal(t, base.UserStatusPending, user.Status)
	assert.Equal(t, base.UserSecurityPasswordOnly, user.SecurityOption)
	assert.Equal(t, now, user.CreatedAt)
	assert.True(t, user.AccessExpireAt.IsZero())
	assert.Empty(t, user.Password)

	other := newSelfSignupUser(&userdto.RequestUserSignupReq{Email: "new@example.com"}, "new", now)
	assert.NotEqual(t, user.ID, other.ID)
	assert.Equal(t, "new", other.FullName, "the username when no name is given")
}

func TestPersistSelfSignupUser(t *testing.T) {
	st := newSignupTest(t, true)
	user := newSelfSignupUser(&userdto.RequestUserSignupReq{Email: "new@example.com", FullName: "New Person"},
		"new", timeutil.NowUTC())

	require.NoError(t, st.uc.persistSelfSignupUser(context.Background(), nil, user))

	require.Len(t, st.svc.persisted, 1)
	persisted := st.svc.persisted[0]
	assert.Equal(t, []*entity.User{user}, persisted.UpsertingUsers)
	assert.Empty(t, persisted.UpsertingAccesses, "no module, no capability")
	assert.Empty(t, persisted.DeletingAccesses)

	require.Len(t, st.audit.entries, 1)
	entry := st.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeUserCreate, entry.Type)
	assert.Equal(t, base.AuditLogSourceAPICreate, entry.Source)
	assert.Equal(t, auditSectionSelfSignup, entry.Section)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Equal(t, user.ID, entry.ObjectID)
	require.NotNil(t, entry.Auth)
	assert.Equal(t, user.ID, entry.Auth.User.ID)
	assert.Contains(t, entry.Detail, `"member"`)
}

func TestSelfSignupLink(t *testing.T) {
	newSignupTest(t, true)

	assert.Equal(t, "https://shellwing.example.com/auth/sign-up?token=tok", selfSignupLink("tok", ""))
	assert.Equal(t,
		"https://shellwing.example.com/auth/sign-up?token=tok&next=%2Faccount%2Ftrial%3Fplan%3Dhive%26x%3D1",
		selfSignupLink("tok", "/account/trial?plan=hive&x=1"))

	// A next that is not a path on the web app never reaches the mail, whatever let it through.
	for _, next := range []string{"https://evil.example.com", "//evil.example.com", `/\evil.example.com`} {
		assert.Equal(t, "https://shellwing.example.com/auth/sign-up?token=tok", selfSignupLink("tok", next), next)
	}
}
