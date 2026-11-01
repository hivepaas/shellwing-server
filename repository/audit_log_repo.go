package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

// AuditLogRepo is deliberately narrower than the other repositories: an audit
// row can be written and read, and removed only by the retention sweep. There is
// no update and no delete-by-id, because a record that can be edited or picked
// off one row at a time is not evidence of anything.
type AuditLogRepo interface {
	// GetByID takes the same scope as List, and hides the row the same way.
	//
	// A nil scope means no scoping - every row is visible. Passing one is what
	// stops an id from being a way around the boundary the listing enforces: the
	// entries record secret reveals, refused attempts and client addresses, so
	// being able to read one by id from outside its scope would make the scoping
	// decorative.
	GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope, id string,
		opts ...bunex.SelectQueryOption) (*entity.AuditLog, error)
	List(ctx context.Context, db database.IDB, scope *entity.ObjectScope, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, auditLog *entity.AuditLog,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, auditLogs []*entity.AuditLog,
		opts ...bunex.InsertQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type auditLogRepo struct {
}

func NewAuditLogRepo() AuditLogRepo {
	return &auditLogRepo{}
}

func (repo *auditLogRepo) GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	id string, opts ...bunex.SelectQueryOption) (*entity.AuditLog, error) {
	theOpts := repo.applyScopeFilter(opts, scope)

	auditLog := &entity.AuditLog{}
	query := db.NewSelect().Model(auditLog).Where("audit_log.id = ?", id)
	query = bunex.ApplySelect(query, theOpts...)

	err := query.Scan(ctx)
	if auditLog == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("AuditLog").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return auditLog, nil
}

func (repo *auditLogRepo) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.AuditLog, *basedto.PagingMeta, error) {
	theOpts := repo.applyScopeFilter(opts, scope)

	var auditLogs []*entity.AuditLog
	query := db.NewSelect().Model(&auditLogs)
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
	return auditLogs, pagingMeta, nil
}

// applyScopeFilter narrows a query to what the given scope may see.
//
// One implementation for both reads. Two would be two chances to disagree about
// what a scope can see, and the half that was more generous would be the one
// somebody found.
func (repo *auditLogRepo) applyScopeFilter(opts []bunex.SelectQueryOption,
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

// applyDirectUserFilter filters logs belong to the user
func (repo *auditLogRepo) applyDirectUserFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("audit_log.object_id = ?", scope.UserID))
	return opts
}

// applyGlobalFilter filters logs belong to global scope
func (repo *auditLogRepo) applyGlobalFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	if scope.NoInherited {
		return append(opts, bunex.SelectWhereIn("audit_log.scope IN (?)", base.ObjectScopeGlobal))
	}
	return opts
}

func (repo *auditLogRepo) Insert(ctx context.Context, db database.IDB, auditLog *entity.AuditLog,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.AuditLog{auditLog}, opts...)
}

func (repo *auditLogRepo) InsertMulti(ctx context.Context, db database.IDB, auditLogs []*entity.AuditLog,
	opts ...bunex.InsertQueryOption) error {
	if len(auditLogs) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&auditLogs)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *auditLogRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.AuditLog)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
