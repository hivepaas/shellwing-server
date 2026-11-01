package emailservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	GetDefaultSystemEmail(ctx context.Context, db database.IDB) (*entity.Setting, error)

	// User emailing
	SendMailPasswordReset(ctx context.Context, db database.IDB, data *EmailDataPasswordReset) error
	SendMailUserInvite(ctx context.Context, db database.IDB, data *EmailDataUserInvite) error
	SendMailSignupVerify(ctx context.Context, db database.IDB, data *EmailDataSignupVerify) error

	// Licensing emailing
}
