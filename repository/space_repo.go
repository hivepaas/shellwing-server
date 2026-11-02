package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type SpaceRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string) (*entity.Space, error)
	// GetPersonal returns nil when the user has no personal space yet.
	GetPersonal(ctx context.Context, db database.IDB, userID string) (*entity.Space, error)
	// InsertIfAbsent makes the space unless the user has their personal space already: two devices asking together
	// make one.
	InsertIfAbsent(ctx context.Context, db database.IDB, space *entity.Space) error
}

type spaceRepo struct {
}

func NewSpaceRepo() SpaceRepo {
	return &spaceRepo{}
}

func (repo *spaceRepo) GetByID(ctx context.Context, db database.IDB, id string) (*entity.Space, error) {
	space := &entity.Space{}
	err := db.NewSelect().Model(space).Where("space.id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("Space").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return space, nil
}

func (repo *spaceRepo) GetPersonal(ctx context.Context, db database.IDB, userID string) (*entity.Space, error) {
	space := &entity.Space{}
	err := db.NewSelect().Model(space).
		Where("space.owner_user_id = ?", userID).
		Where("space.type = 'personal'").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return space, nil
}

func (repo *spaceRepo) InsertIfAbsent(ctx context.Context, db database.IDB, space *entity.Space) error {
	_, err := db.NewInsert().Model(space).
		On("CONFLICT (owner_user_id) WHERE type = 'personal' AND deleted_at IS NULL DO NOTHING").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
