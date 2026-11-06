package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// SyncRepo holds the metadata of synced files: never their content.
type SyncRepo interface {
	EnsureSpace(ctx context.Context, db database.IDB, spaceID string) error
	GetSpace(ctx context.Context, db database.IDB, spaceID string) (*entity.SyncSpace, error)
	// LockSpace reads the totals and holds the row until the transaction ends, which puts a space's writes in one
	// line: revisions commit in order, so a client that has seen revision N has seen everything up to it.
	LockSpace(ctx context.Context, db database.IDB, spaceID string) (*entity.SyncSpace, error)
	UpdateSpace(ctx context.Context, db database.IDB, space *entity.SyncSpace) error

	// GetObject returns nil when there has never been such a file.
	GetObject(ctx context.Context, db database.IDB, spaceID, name string) (*entity.SyncObject, error)
	UpsertObject(ctx context.Context, db database.IDB, object *entity.SyncObject) error
	// ListChanges returns the lines changed after the revision, oldest first.
	ListChanges(ctx context.Context, db database.IDB, spaceID string, since int64, limit int) ([]*entity.SyncObject, error)
	// CurrentETags returns the ETag of each live file among the names.
	CurrentETags(ctx context.Context, db database.IDB, spaceID string, names []string) (map[string]string, error)

	// AddTrash notes content that may go after the delay. Noted twice, the later time counts.
	AddTrash(ctx context.Context, db database.IDB, spaceID, name, etag string, deleteAfter time.Time) error
	ListDueTrash(ctx context.Context, db database.IDB, now time.Time, limit int) ([]*entity.SyncTrash, error)
	DeleteTrash(ctx context.Context, db database.IDB, item *entity.SyncTrash) error
	// TrashAll removes every file of a space, noting their content to go after the delay. The space is not changed:
	// nobody syncs it any more.
	TrashAll(ctx context.Context, db database.IDB, spaceID string, deleteAfter time.Time) error
	// TrashedETags returns the versions of the names noted in the trash, as name to ETags.
	TrashedETags(ctx context.Context, db database.IDB, spaceID string, names []string) (map[string][]string, error)
}

type syncRepo struct {
}

func NewSyncRepo() SyncRepo {
	return &syncRepo{}
}

func (repo *syncRepo) EnsureSpace(ctx context.Context, db database.IDB, spaceID string) error {
	_, err := db.NewInsert().Model(&entity.SyncSpace{SpaceID: spaceID}).
		On("CONFLICT (space_id) DO NOTHING").Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) GetSpace(ctx context.Context, db database.IDB, spaceID string) (*entity.SyncSpace, error) {
	return repo.getSpace(ctx, db.NewSelect(), spaceID)
}

func (repo *syncRepo) LockSpace(ctx context.Context, db database.IDB, spaceID string) (*entity.SyncSpace, error) {
	return repo.getSpace(ctx, db.NewSelect().For("UPDATE"), spaceID)
}

func (repo *syncRepo) getSpace(ctx context.Context, query *bun.SelectQuery, spaceID string) (*entity.SyncSpace, error) {
	space := &entity.SyncSpace{}
	err := query.Model(space).Where("sync_space.space_id = ?", spaceID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("Space").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return space, nil
}

func (repo *syncRepo) UpdateSpace(ctx context.Context, db database.IDB, space *entity.SyncSpace) error {
	_, err := db.NewUpdate().Model(space).WherePK().Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) GetObject(ctx context.Context, db database.IDB,
	spaceID, name string) (*entity.SyncObject, error) {
	object := &entity.SyncObject{}
	err := db.NewSelect().Model(object).
		Where("sync_object.space_id = ?", spaceID).Where("sync_object.name = ?", name).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return object, nil
}

func (repo *syncRepo) UpsertObject(ctx context.Context, db database.IDB, object *entity.SyncObject) error {
	_, err := db.NewInsert().Model(object).
		On("CONFLICT (space_id, name) DO UPDATE").
		Set("etag = EXCLUDED.etag").
		Set("size = EXCLUDED.size").
		Set("kind = EXCLUDED.kind").
		Set("tombstone = EXCLUDED.tombstone").
		Set("revision = EXCLUDED.revision").
		Set("updated_at = EXCLUDED.updated_at").
		Set("deleted_at = EXCLUDED.deleted_at").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) ListChanges(ctx context.Context, db database.IDB, spaceID string, since int64,
	limit int) ([]*entity.SyncObject, error) {
	var objects []*entity.SyncObject
	query := db.NewSelect().Model(&objects).
		Where("sync_object.space_id = ?", spaceID).
		Where("sync_object.revision > ?", since).
		Order("sync_object.revision ASC").
		Limit(int64(limit))
	// A client starting from nothing has no copy of what was removed.
	if since == 0 {
		query = query.Where("sync_object.deleted_at IS NULL")
	}
	if err := query.Scan(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return objects, nil
}

func (repo *syncRepo) CurrentETags(ctx context.Context, db database.IDB, spaceID string,
	names []string) (map[string]string, error) {
	var objects []*entity.SyncObject
	err := db.NewSelect().Model(&objects).
		Column("name", "etag").
		Where("sync_object.space_id = ?", spaceID).
		Where("sync_object.name IN (?)", bun.List(names)).
		Where("sync_object.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	etags := make(map[string]string, len(objects))
	for _, object := range objects {
		etags[object.Name] = object.ETag
	}
	return etags, nil
}

func (repo *syncRepo) AddTrash(ctx context.Context, db database.IDB, spaceID, name, etag string,
	deleteAfter time.Time) error {
	item := &entity.SyncTrash{SpaceID: spaceID, Name: name, ETag: etag, DeleteAfter: deleteAfter}
	_, err := db.NewInsert().Model(item).
		On("CONFLICT (space_id, name, etag) DO UPDATE").
		Set("delete_after = EXCLUDED.delete_after").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) TrashAll(ctx context.Context, db database.IDB, spaceID string, deleteAfter time.Time) error {
	_, err := db.NewRaw(`
INSERT INTO sync_trash (space_id, name, etag, delete_after)
SELECT space_id, name, etag, ? FROM sync_objects WHERE space_id = ? AND deleted_at IS NULL
ON CONFLICT (space_id, name, etag) DO UPDATE SET delete_after = EXCLUDED.delete_after`,
		deleteAfter, spaceID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = db.NewUpdate().Model((*entity.SyncObject)(nil)).
		Set("deleted_at = CURRENT_TIMESTAMP").Set("updated_at = CURRENT_TIMESTAMP").
		Where("space_id = ?", spaceID).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) ListDueTrash(ctx context.Context, db database.IDB, now time.Time,
	limit int) ([]*entity.SyncTrash, error) {
	var items []*entity.SyncTrash
	err := db.NewSelect().Model(&items).
		Where("sync_trash.delete_after <= ?", now).
		Order("sync_trash.delete_after ASC").
		Limit(int64(limit)).
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return items, nil
}

func (repo *syncRepo) DeleteTrash(ctx context.Context, db database.IDB, item *entity.SyncTrash) error {
	_, err := db.NewDelete().Model(item).WherePK().Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *syncRepo) TrashedETags(ctx context.Context, db database.IDB, spaceID string,
	names []string) (map[string][]string, error) {
	var items []*entity.SyncTrash
	err := db.NewSelect().Model(&items).
		Where("sync_trash.space_id = ?", spaceID).
		Where("sync_trash.name IN (?)", bun.List(names)).
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	etags := make(map[string][]string, len(items))
	for _, item := range items {
		etags[item.Name] = append(etags[item.Name], item.ETag)
	}
	return etags, nil
}
