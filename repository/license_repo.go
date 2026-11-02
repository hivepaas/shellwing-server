package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// userSeatLockKey names the advisory lock over the count of users. Any number will do that nothing else uses.
const userSeatLockKey = 7301002

type LicenseRepo interface {
	// InsertIfAbsent makes the row unless it is there: two instances starting together make one.
	InsertIfAbsent(ctx context.Context, db database.IDB, state *entity.LicenseState) error
	// Get returns nil when the row is not there yet.
	Get(ctx context.Context, db database.IDB) (*entity.LicenseState, error)
	Update(ctx context.Context, db database.IDB, state *entity.LicenseState) error

	// LockUserSeats holds the count of users until the transaction ends: of two creations at once, the second waits
	// for the first to commit, and counts it. Outside a transaction it would be let go at once.
	LockUserSeats(ctx context.Context, db database.IDB) error
	// CountUserSeats counts the users who hold a seat: active, or invited and not there yet. A disabled user holds none.
	CountUserSeats(ctx context.Context, db database.IDB) (int, error)
}

type licenseRepo struct {
}

func NewLicenseRepo() LicenseRepo {
	return &licenseRepo{}
}

func (repo *licenseRepo) InsertIfAbsent(ctx context.Context, db database.IDB, state *entity.LicenseState) error {
	_, err := db.NewInsert().Model(state).On("CONFLICT (id) DO NOTHING").Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *licenseRepo) Get(ctx context.Context, db database.IDB) (*entity.LicenseState, error) {
	state := &entity.LicenseState{}
	err := db.NewSelect().Model(state).Where("license_state.id = 1").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return state, nil
}

func (repo *licenseRepo) Update(ctx context.Context, db database.IDB, state *entity.LicenseState) error {
	_, err := db.NewUpdate().Model(state).Column("key_text", "updated_at").WherePK().Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *licenseRepo) LockUserSeats(ctx context.Context, db database.IDB) error {
	_, err := db.ExecContext(ctx, "SELECT pg_advisory_xact_lock(?)", userSeatLockKey)
	return hperrors.Wrap(err)
}

func (repo *licenseRepo) CountUserSeats(ctx context.Context, db database.IDB) (int, error) {
	count, err := db.NewSelect().Model((*entity.User)(nil)).
		Where("status IN (?, ?)", base.UserStatusActive, base.UserStatusPending).
		Count(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int(count), nil
}
