package userservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/entity/appentity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

type Service interface {
	InitAdminUser(ctx context.Context, db database.IDB) (err error)

	LoadUser(ctx context.Context, db database.IDB, userID string, errorIfUnavailable bool,
		loadOpts ...bunex.SelectQueryOption) (*entity.User, error)
	LoadUsers(ctx context.Context, db database.IDB, userIDs []string, errorIfUnavailable bool,
		loadOpts ...bunex.SelectQueryOption) (map[string]*entity.User, error)
	LoadUsersSkipMissing(ctx context.Context, db database.IDB, userIDs []string, errorIfUnavailable bool,
		loadOpts ...bunex.SelectQueryOption) (map[string]*entity.User, error)
	LoadUsersCustomConds(ctx context.Context, db database.IDB, errorIfUnavailable bool,
		loadOpts ...bunex.SelectQueryOption) (map[string]*entity.User, error)

	PersistUserData(ctx context.Context, db database.IDB, persistingData *PersistingUserData) error
	DeleteUser(ctx context.Context, db database.IDB, user *entity.User) error

	ChangePassword(user *entity.User, newPassword, currPassword string) error
	VerifyPassword(user *entity.User, password string) error

	GenerateMFAToken(userID string, mfaType base.MFAType, trustedDeviceID string) (string, error)
	ParseMFAToken(token string) (*appentity.MFATokenClaims, error)
	GenerateMFATotpSetupToken(userID string, toptSecret string) (string, error)
	ParseMFATotpSetupToken(token string) (*appentity.MFATotpSetupTokenClaims, error)
	GenerateUserInviteToken(userID string) (string, error)
	ParseUserInviteToken(token string) (*appentity.UserInviteTokenClaims, error)
	GeneratePasswordResetToken(userID string) (string, error)
	ParsePasswordResetToken(token string) (*appentity.PasswordResetTokenClaims, error)
	GenerateSSOSignupToken(claims *appentity.SSOSignupTokenClaims) (string, error)
	ParseSSOSignupToken(token string) (*appentity.SSOSignupTokenClaims, error)

	// Project users
	LoadNotificationUsers(ctx context.Context, db database.IDB, loadAdmins bool) (map[string]*entity.User, error)
}
