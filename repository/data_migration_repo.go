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

type DataMigrationRepo interface {
	GetLatest(ctx context.Context, db database.IDB,
		opts ...bunex.SelectQueryOption) (*entity.DataMigration, error)

	Insert(ctx context.Context, db database.IDB, dataMigration *entity.DataMigration,
		opts ...bunex.InsertQueryOption) error
}

type dataMigrationRepo struct {
}

func NewDataMigrationRepo() DataMigrationRepo {
	return &dataMigrationRepo{}
}

func (repo *dataMigrationRepo) GetLatest(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) (*entity.DataMigration, error) {
	dataMigration := &entity.DataMigration{}
	query := db.NewSelect().Model(dataMigration).Order("id DESC")
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if dataMigration == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("DataMigration").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return dataMigration, nil
}

func (repo *dataMigrationRepo) Insert(ctx context.Context, db database.IDB, dataMigration *entity.DataMigration,
	opts ...bunex.InsertQueryOption) error {
	query := db.NewInsert().Model(dataMigration)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
