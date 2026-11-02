package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type DeviceRepo interface {
	Insert(ctx context.Context, db database.IDB, device *entity.Device) error
	Update(ctx context.Context, db database.IDB, device *entity.Device) error
	// GetByRefreshTokenHash returns nil when no device has the token.
	GetByRefreshTokenHash(ctx context.Context, db database.IDB, hash string) (*entity.Device, error)
	// GetOfUser returns nil when the user has no such device.
	GetOfUser(ctx context.Context, db database.IDB, userID, id string) (*entity.Device, error)
	// ListOfUser returns the devices not revoked, the latest seen first.
	ListOfUser(ctx context.Context, db database.IDB, userID string) ([]*entity.Device, error)
	RevokeAllOfUser(ctx context.Context, db database.IDB, userID string, now time.Time) error
}

type deviceRepo struct {
}

func NewDeviceRepo() DeviceRepo {
	return &deviceRepo{}
}

func (repo *deviceRepo) Insert(ctx context.Context, db database.IDB, device *entity.Device) error {
	_, err := db.NewInsert().Model(device).Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *deviceRepo) Update(ctx context.Context, db database.IDB, device *entity.Device) error {
	_, err := db.NewUpdate().Model(device).WherePK().Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *deviceRepo) GetByRefreshTokenHash(ctx context.Context, db database.IDB,
	hash string) (*entity.Device, error) {
	device := &entity.Device{}
	err := db.NewSelect().Model(device).Where("device.refresh_token_hash = ?", hash).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return device, nil
}

func (repo *deviceRepo) GetOfUser(ctx context.Context, db database.IDB, userID, id string) (*entity.Device, error) {
	device := &entity.Device{}
	err := db.NewSelect().Model(device).Where("device.id = ?", id).Where("device.user_id = ?", userID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return device, nil
}

func (repo *deviceRepo) ListOfUser(ctx context.Context, db database.IDB, userID string) ([]*entity.Device, error) {
	var devices []*entity.Device
	err := db.NewSelect().Model(&devices).
		Where("device.user_id = ?", userID).
		Where("device.revoked_at IS NULL").
		OrderExpr("device.last_seen_at DESC NULLS LAST").
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return devices, nil
}

func (repo *deviceRepo) RevokeAllOfUser(ctx context.Context, db database.IDB, userID string, now time.Time) error {
	_, err := db.NewUpdate().Model((*entity.Device)(nil)).
		Set("revoked_at = ?", now).
		Set("refresh_token_hash = NULL").
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)
	return hperrors.Wrap(err)
}
