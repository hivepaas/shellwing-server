package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

type SystemStatusRepo interface {
	Get(ctx context.Context, db database.IDB, opts ...bunex.SelectQueryOption) (*entity.SystemStatus, error)

	Upsert(ctx context.Context, db database.IDB, systemStatus *entity.SystemStatus, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
}

type systemStatusRepo struct {
}

func NewSystemStatusRepo() SystemStatusRepo {
	return &systemStatusRepo{}
}

func (repo *systemStatusRepo) Get(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) (*entity.SystemStatus, error) {
	systemStatus := &entity.SystemStatus{}
	query := db.NewSelect().Model(systemStatus).Where("system_status.id = ?", 1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if systemStatus == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("SystemStatus").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return systemStatus, nil
}

func (repo *systemStatusRepo) Upsert(ctx context.Context, db database.IDB, systemStatus *entity.SystemStatus,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	query := db.NewInsert().Model(systemStatus)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
