package taskserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

const (
	maxTryLock = 3
)

func (s *service) CreateDBLock(
	ctx context.Context,
	db database.Tx,
	id string,
	selectFor string,
) (*entity.Lock, error) {
	return s.tryCreateDBLock(ctx, db, id, selectFor, 1)
}

func (s *service) tryCreateDBLock(
	ctx context.Context,
	db database.Tx,
	id string,
	selectFor string,
	try int,
) (*entity.Lock, error) {
	if try >= maxTryLock {
		return nil, hperrors.Wrap(hperrors.ErrActionFailed)
	}
	if selectFor == "" {
		selectFor = "UPDATE"
	}
	lock, err := s.lockRepo.GetByID(ctx, db, id,
		bunex.SelectFor(selectFor),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}

	if lock != nil {
		return lock, nil
	}

	// Insert a new row for the lock using the default DB (not the transaction one)
	lock = &entity.Lock{ID: id}
	err = s.lockRepo.Upsert(ctx, s.db, lock, entity.LockUpsertingConflictCols, entity.LockUpsertingUpdateCols)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	try++
	return s.tryCreateDBLock(ctx, db, id, selectFor, try)
}
