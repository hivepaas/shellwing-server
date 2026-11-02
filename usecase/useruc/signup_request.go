package useruc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/reqinfo"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/emailservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

// How often sign-up may be asked for. Each request can send a mail, so these keep the form from being
// used to mail somebody over and over, or to spray an address list; the per-email limit also bounds how
// many links one inbox can be sent, whoever asks.
const (
	signupRequestsPerIP    = 5
	signupRequestsPerEmail = 3
	signupRequestWindow    = time.Hour
)

// signupRequestAction is what a sign-up request does, by what already has the address.
type signupRequestAction int

const (
	// signupRequestCreate makes a pending account and mails its link: nobody has the address.
	signupRequestCreate signupRequestAction = iota + 1
	// signupRequestResend mails a new link to an account nobody has finished, whether a sign-up or an
	// admin's invitation made it.
	signupRequestResend
	// signupRequestIgnore sends nothing: the address already has an account in use.
	signupRequestIgnore
)

// signupRequestActionFor decides what a sign-up request does for the account that has its address, nil when
// none does.
func signupRequestActionFor(user *entity.User) signupRequestAction {
	if user == nil {
		return signupRequestCreate
	}
	switch user.Status {
	case base.UserStatusPending:
		return signupRequestResend
	case base.UserStatusActive, base.UserStatusDisabled:
		return signupRequestIgnore
	default:
		// A status this code does not know is somebody's account all the same.
		return signupRequestIgnore
	}
}

// RequestUserSignup lets anybody ask for an account with an email address. The link mailed to the address
// is what proves it is theirs, and finishing the account goes through CompleteUserSignup like an invitation.
//
// The answer is the same empty success whatever the address has - a new account, an unfinished one, or
// one in use - so the form cannot be used to find out who has an account here.
func (uc *UC) RequestUserSignup(
	ctx context.Context,
	req *userdto.RequestUserSignupReq,
) (*userdto.RequestUserSignupResp, error) {
	if !config.Current().Users.SelfSignupEnabled {
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowedByAdmin).
			WithMsgLog("self sign-up is disabled (users.self_signup_enabled)")
	}
	if err := uc.limitSignupRequests(ctx, req.Email); err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Whatever the address has, so that a server with no mail set up answers every address alike.
	systemEmail, err := uc.signupSystemEmail(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	user, err := uc.userRepo.GetByEmail(ctx, uc.db, req.Email,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		if !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		user = nil
	}

	switch signupRequestActionFor(user) {
	case signupRequestIgnore:
		// Kept as a refusal against the account, with the address the request came from: one account's
		// address tried here again and again is somebody looking for accounts, or for a way into one.
		err = uc.recordUserChangeDenied(ctx, uc.db, nil, auditSectionSelfSignup, user, auditReasonAccountExists)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return &userdto.RequestUserSignupResp{}, nil
	case signupRequestCreate:
		user, err = uc.createSelfSignupUser(ctx, req)
		if errors.Is(err, hperrors.ErrPlanLimitReached) {
			// The same empty answer, with nothing sent: telling only a new address so would tell which ones are
			// taken. An admin finds no new account, and the server's limit says why.
			return &userdto.RequestUserSignupResp{}, nil
		}
		if bunex.IsErrorUniqueViolation(err) {
			// A concurrent request took the address, or the username made from it, first. Nothing was
			// written here; asking again sends a link.
			return &userdto.RequestUserSignupResp{}, nil
		}
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	case signupRequestResend:
		if err = uc.recordSignupLinkResent(ctx, uc.db, user); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if err = uc.sendSignupLink(ctx, systemEmail, user, req.Next); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &userdto.RequestUserSignupResp{}, nil
}

// limitSignupRequests counts the request against the address it came from and against the email it names,
// and refuses it once either is over its limit. The address is counted first, so a caller already over its
// own limit no longer uses up the email's.
func (uc *UC) limitSignupRequests(ctx context.Context, email string) error {
	clientIP := gofn.Coalesce(reqinfo.ClientIPFrom(ctx), "unknown")
	count, err := uc.rateLimitRepo.Hit(ctx, "signup-request:ip:"+clientIP, signupRequestWindow)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count > signupRequestsPerIP {
		return hperrors.Wrap(hperrors.ErrTooManyRequests).
			WithMsgLog("sign-up requests from %s over %d an hour", clientIP, signupRequestsPerIP)
	}

	// The email is hashed so that the cache holds no address somebody merely typed.
	sum := sha256.Sum256([]byte(email))
	count, err = uc.rateLimitRepo.Hit(ctx, "signup-request:email:"+hex.EncodeToString(sum[:]),
		signupRequestWindow)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count > signupRequestsPerEmail {
		return hperrors.Wrap(hperrors.ErrTooManyRequests).
			WithMsgLog("sign-up requests for one email over %d an hour", signupRequestsPerEmail)
	}
	return nil
}

func (uc *UC) signupSystemEmail(ctx context.Context) (*entity.Email, error) {
	emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrActionNotAllowedByAdmin).WithCause(err).
			WithMsgLog("self sign-up needs a default system email")
	}
	email, err := emailSetting.AsEmail()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return email, nil
}

func (uc *UC) createSelfSignupUser(
	ctx context.Context,
	req *userdto.RequestUserSignupReq,
) (user *entity.User, err error) {
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if e := uc.licenseService.CheckUserSeat(ctx, db); e != nil {
			return hperrors.Wrap(e)
		}
		username, e := userservice.UsernameForEmail(ctx, db, uc.userRepo, req.Email)
		if e != nil {
			return hperrors.Wrap(e)
		}
		user = newSelfSignupUser(req, username, timeutil.NowUTC())
		return uc.persistSelfSignupUser(ctx, db, user)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return user, nil
}

// newSelfSignupUser is the account a sign-up request makes: a member with no module and no capability,
// signing in with a password, and pending until the link is used. What it may do in the server is its own
// licenses, which need no grant.
func newSelfSignupUser(req *userdto.RequestUserSignupReq, username string, timeNow time.Time) *entity.User {
	return &entity.User{
		ID:             gofn.Must(ulid.NewStringULID()),
		Username:       username,
		Email:          req.Email,
		FullName:       gofn.Coalesce(req.FullName, username),
		Role:           base.UserRoleMember,
		Status:         base.UserStatusPending,
		SecurityOption: base.UserSecurityPasswordOnly,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
	}
}

func (uc *UC) persistSelfSignupUser(ctx context.Context, db database.IDB, user *entity.User) error {
	err := uc.userService.PersistUserData(ctx, db, &userservice.PersistingUserData{
		UpsertingUsers: []*entity.User{user},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Attributed to the account, for want of anybody else behind the request; the section says the
	// address was not yet shown to be theirs. The account can do nothing until the link is used, and
	// finishing it is recorded then, under auditSectionSignup.
	return uc.recordUserChange(ctx, db, authOfUser(user), base.AuditLogTypeUserCreate,
		auditSectionSelfSignup, user, auditdetail.New().
			Set("role", user.Role).
			Set("securityOption", user.SecurityOption))
}

// recordSignupLinkResent records a new sign-up link going out for an unfinished account. Like the
// forgot-password form's entry, it names no actor: the request gives an address and nothing else, and
// naming the account would have the entry say the account asked.
func (uc *UC) recordSignupLinkResent(ctx context.Context, db database.IDB, user *entity.User) error {
	err := uc.auditService.Record(ctx, db, &auditservice.Entry{
		Type:     base.AuditLogTypeUserUpdate,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.ID,
		Source:   base.AuditLogSourceAPIUpdate,
		Section:  auditSectionSelfSignup,
		Result:   base.AuditLogResultAllowed,
		ResType:  base.ResourceTypeUser,
		ResID:    user.ID,
		ResName:  user.Username,
		Detail:   auditdetail.New().Set("linkResent", true).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) sendSignupLink(
	ctx context.Context,
	systemEmail *entity.Email,
	user *entity.User,
	next string,
) error {
	token, err := uc.userService.GenerateUserInviteToken(user.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.emailService.SendMailSignupVerify(ctx, uc.db, &emailservice.EmailDataSignupVerify{
		BaseTemplateData: emailservice.BaseTemplateData{
			Email:      systemEmail,
			Recipients: []string{user.Email},
		},
		SignupLink: selfSignupLink(token, next),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// selfSignupLink is the invitation's sign-up page, plus the page to open once the account is finished.
// next was validated with the request; it is checked again because it ends up in a link the server mails
// under its own name, and one that is not a path on the web app is dropped rather than sent.
func selfSignupLink(token, next string) string {
	link := config.Current().DashboardUserSignupURL(token)
	if next != "" && userdto.IsSafeNextPath(next) {
		link += "&next=" + url.QueryEscape(next)
	}
	return link
}
