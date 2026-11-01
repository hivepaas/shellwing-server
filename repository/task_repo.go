package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

type TaskRepo interface {
	GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.TaskType, id string,
		opts ...bunex.SelectQueryOption) (*entity.Task, error)
	List(ctx context.Context, db database.IDB, scope *entity.ObjectScope, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error)
	ListByTarget(ctx context.Context, db database.IDB, targetID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.Task, error)

	Insert(ctx context.Context, db database.IDB, task *entity.Task,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, task *entity.Task, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, task *entity.Task,
		opts ...bunex.UpdateQueryOption) error

	DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteAllByUsers(ctx context.Context, db database.IDB, userIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB,
		opts ...bunex.DeleteQueryOption) error
}

type taskRepo struct {
}

func NewTaskRepo() TaskRepo {
	return &taskRepo{}
}

func (repo *taskRepo) GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.TaskType, id string, opts ...bunex.SelectQueryOption) (*entity.Task, error) {
	theOpts := repo.applyScopeFilter(opts, scope)

	task := &entity.Task{}
	query := db.NewSelect().Model(task).Where("task.id = ?", id)
	if typ != "" {
		query = query.Where("task.type = ?", typ)
	}
	query = bunex.ApplySelect(query, theOpts...)

	err := query.Scan(ctx)
	if task == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("Task").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

func (repo *taskRepo) ListByTarget(ctx context.Context, db database.IDB, targetID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks)
	if targetID != "" {
		query = query.Where("task.target_id = ?", targetID)
	}
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}
	return tasks, pagingMeta, nil
}

func (repo *taskRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.Task, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks).Where("task.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return tasks, nil
}

func (repo *taskRepo) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	theOpts := repo.applyScopeFilter(opts, scope)

	var tasks []*entity.Task
	query := db.NewSelect().Model(&tasks)
	query = bunex.ApplySelect(query, theOpts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}
	return tasks, pagingMeta, nil
}

// applyScopeFilter narrows a query to what the given scope may see.
//
// One implementation for both reads. Two would be two chances to disagree about
// what a scope can see, and the half that was more generous would be the one
// somebody found.
func (repo *taskRepo) applyScopeFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope == nil {
		return opts
	}
	switch scope.ScopeType {
	case base.ObjectScopeUser:
		return repo.applyDirectUserFilter(opts, scope)
	case base.ObjectScopeGlobal:
		return repo.applyGlobalFilter(opts, scope)
	default:
		return opts
	}
}

// applyDirectUserFilter filters tasks belong to the user
func (repo *taskRepo) applyDirectUserFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("task.object_id = ?", scope.UserID))
	return opts
}

// applyGlobalFilter filters tasks belong to global scope
func (repo *taskRepo) applyGlobalFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts, bunex.SelectWhereIn("task.scope IN (?)", base.ObjectScopeGlobal))
	}
	return opts
}

func (repo *taskRepo) Insert(ctx context.Context, db database.IDB, task *entity.Task,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.Task{task}, opts...)
}

func (repo *taskRepo) InsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	opts ...bunex.InsertQueryOption) error {
	if len(tasks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&tasks)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) Upsert(ctx context.Context, db database.IDB, task *entity.Task,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Task{task}, conflictCols, updateCols, opts...)
}

func (repo *taskRepo) UpsertMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(tasks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&tasks)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) Update(ctx context.Context, db database.IDB, task *entity.Task,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(task).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// NOTE: this UpdateMulti may not work properly with JSON columns defined as string in the entity struct
func (repo *taskRepo) UpdateMulti(ctx context.Context, db database.IDB, tasks []*entity.Task,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(&tasks).Bulk()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(ids) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.id IN (?)", bun.List(ids))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteAllByUsers(ctx context.Context, db database.IDB, userIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(userIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).
		Where("task.target_id IN (?)", bun.List(userIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Task)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
