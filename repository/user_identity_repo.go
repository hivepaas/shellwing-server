package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// UserIdentityRepo stores each person's published identity: one per person.
type UserIdentityRepo interface {
	// Get returns nil when the person has published none.
	Get(ctx context.Context, db database.IDB, userID string) (*entity.UserIdentity, error)
	// Put saves the person's identity, in place of one they had.
	Put(ctx context.Context, db database.IDB, identity *entity.UserIdentity) error
}

type userIdentityRepo struct {
}

func NewUserIdentityRepo() UserIdentityRepo {
	return &userIdentityRepo{}
}

func (repo *userIdentityRepo) Get(ctx context.Context, db database.IDB, userID string) (*entity.UserIdentity, error) {
	identity := &entity.UserIdentity{}
	err := db.NewSelect().Model(identity).Where("user_id = ?", userID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return identity, nil
}

func (repo *userIdentityRepo) Put(ctx context.Context, db database.IDB, identity *entity.UserIdentity) error {
	_, err := db.NewInsert().Model(identity).
		On("CONFLICT (user_id) DO UPDATE").
		Set("x25519_public = EXCLUDED.x25519_public").
		Set("ed25519_public = EXCLUDED.ed25519_public").
		Set("signature = EXCLUDED.signature").
		Set("fingerprint = EXCLUDED.fingerprint").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
