package userserviceimpl

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

func (s *service) InitAdminUser(
	ctx context.Context,
	db database.IDB,
) (err error) {
	accCfg := &config.Current().Users.Admin
	email := accCfg.Email
	password := accCfg.Password
	if email == "" || password == "" {
		return hperrors.NewMissing("Email or password is missing")
	}
	username := gofn.Coalesce(accCfg.Username, strings.Split(email, "@")[0])

	timeNow := timeutil.NowUTC()
	user := &entity.User{
		ID:             gofn.Must(ulid.NewStringULID()),
		Email:          email,
		Username:       username,
		Role:           base.UserRoleAdmin,
		Status:         base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
		CreatedAt:      timeNow,
		UpdatedAt:      timeNow,
	}

	user.Password, err = s.createPasswordHash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	err = s.userRepo.Upsert(ctx, db, user, entity.UserUpsertingConflictCols, entity.UserUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
