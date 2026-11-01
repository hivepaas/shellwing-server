package userservice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
)

// UsernameForEmail is a new account's first username: the email's local part, with a random number added when
// that is taken. The person may choose another later.
func UsernameForEmail(ctx context.Context, db database.IDB, userRepo repository.UserRepo, email string) (
	string, error) {
	username := strings.SplitN(email, "@", 2)[0] //nolint:mnd
	conflictUser, err := userRepo.GetByUsername(ctx, db, username)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return "", hperrors.Wrap(err)
	}
	if conflictUser != nil {
		username += fmt.Sprintf("-%d", 1000+rand.Intn(9000)) //nolint
	}
	return username, nil
}
