package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// CollectionRepo stores an organization's collections, who is granted what in them, and their keys as wrapped for
// each person. Every call names the organization: one's collection is never reached through another's.
type CollectionRepo interface {
	// List returns the organization's collections, by name.
	List(ctx context.Context, db database.IDB, orgID string) ([]*entity.Collection, error)
	// Get returns nil when the organization has no such collection.
	Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.Collection, error)
	// Lock is Get, holding the collection's row until the transaction ends.
	Lock(ctx context.Context, db database.IDB, orgID, id string) (*entity.Collection, error)
	// NameTaken says whether another collection of the organization has the name, whatever its case.
	NameTaken(ctx context.Context, db database.IDB, orgID, name, exceptID string) (bool, error)
	Insert(ctx context.Context, db database.IDB, collection *entity.Collection) error
	// Update saves a collection's name and description.
	Update(ctx context.Context, db database.IDB, collection *entity.Collection) error
	SetKeyVersion(ctx context.Context, db database.IDB, orgID, id string, version int) error
	Delete(ctx context.Context, db database.IDB, orgID, id string) error

	// ListGrantees returns a collection's grants, people first then groups, by name.
	ListGrantees(ctx context.Context, db database.IDB, orgID, collectionID string) ([]*entity.CollectionGrantee, error)
	// SetGrants makes these the collection's grants, and no others.
	SetGrants(ctx context.Context, db database.IDB, orgID, collectionID string, grants []*entity.CollectionGrant) error
	// RemoveUserGrants takes away the grants made to a person by name, in every collection of the organization.
	RemoveUserGrants(ctx context.Context, db database.IDB, orgID, userID string) error
	// Entitlements returns, for each collection and person of the organization who is not removed from it, the most
	// they may do: Manage for an owner, else the most of their grant and their groups'. People with nothing are left
	// out. An empty collectionID or userID does not filter.
	Entitlements(ctx context.Context, db database.IDB, orgID, collectionID, userID string) (
		[]*entity.CollectionEntitlement, error)

	// ListWraps returns the person's wraps of a collection's key, newest version first.
	ListWraps(ctx context.Context, db database.IDB, collectionID, userID string) ([]*entity.CollectionKeyWrap, error)
	InsertWraps(ctx context.Context, db database.IDB, wraps []*entity.CollectionKeyWrap) error
	// KeyHolders returns who holds the current key of the organization's collections: of one, unless collectionID is
	// empty.
	KeyHolders(ctx context.Context, db database.IDB, orgID, collectionID string) ([]*entity.CollectionKeyHolder, error)
	// DeleteUserWraps drops every wrap made for the person: an identity that replaces theirs cannot open them.
	DeleteUserWraps(ctx context.Context, db database.IDB, userID string) (int, error)
}

type collectionRepo struct {
}

func NewCollectionRepo() CollectionRepo {
	return &collectionRepo{}
}

func (repo *collectionRepo) List(ctx context.Context, db database.IDB, orgID string) ([]*entity.Collection, error) {
	collections := []*entity.Collection{}
	err := db.NewSelect().Model(&collections).Where("org_id = ?", orgID).OrderExpr("lower(name)").Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return collections, nil
}

func (repo *collectionRepo) get(ctx context.Context, query *bun.SelectQuery, orgID, id string) (
	*entity.Collection, error) {
	collection := &entity.Collection{}
	err := query.Model(collection).Where("org_id = ?", orgID).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return collection, nil
}

func (repo *collectionRepo) Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.Collection, error) {
	return repo.get(ctx, db.NewSelect(), orgID, id)
}

func (repo *collectionRepo) Lock(ctx context.Context, db database.IDB, orgID, id string) (*entity.Collection, error) {
	return repo.get(ctx, db.NewSelect().For("UPDATE"), orgID, id)
}

func (repo *collectionRepo) NameTaken(ctx context.Context, db database.IDB, orgID, name, exceptID string) (
	bool, error) {
	query := db.NewSelect().Model((*entity.Collection)(nil)).
		Where("org_id = ?", orgID).Where("lower(name) = lower(?)", name)
	if exceptID != "" {
		query = query.Where("id <> ?", exceptID)
	}
	taken, err := query.Exists(ctx)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return taken, nil
}

func (repo *collectionRepo) Insert(ctx context.Context, db database.IDB, collection *entity.Collection) error {
	if _, err := db.NewInsert().Model(collection).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) Update(ctx context.Context, db database.IDB, collection *entity.Collection) error {
	_, err := db.NewUpdate().Model(collection).
		Column("name", "description").
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("org_id = ?", collection.OrgID).WherePK().Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) SetKeyVersion(ctx context.Context, db database.IDB, orgID, id string, version int) error {
	_, err := db.NewUpdate().Model((*entity.Collection)(nil)).
		Set("key_version = ?", version).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("org_id = ?", orgID).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) Delete(ctx context.Context, db database.IDB, orgID, id string) error {
	_, err := db.NewDelete().Model((*entity.Collection)(nil)).
		Where("org_id = ?", orgID).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) ListGrantees(ctx context.Context, db database.IDB, orgID, collectionID string) (
	[]*entity.CollectionGrantee, error) {
	grantees := []*entity.CollectionGrantee{}
	err := db.NewSelect().Model(&grantees).
		ColumnExpr("collection_grant.*").
		ColumnExpr("COALESCE(u.username, '') AS username, COALESCE(u.full_name, '') AS full_name").
		ColumnExpr("COALESCE(g.name, '') AS group_name").
		Join("LEFT JOIN users AS u ON u.id = collection_grant.user_id").
		Join("LEFT JOIN org_groups AS g ON g.id = collection_grant.group_id").
		Where("collection_grant.org_id = ?", orgID).
		Where("collection_grant.collection_id = ?", collectionID).
		OrderExpr("collection_grant.group_id IS NOT NULL, lower(COALESCE(u.username, g.name))").
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return grantees, nil
}

func (repo *collectionRepo) SetGrants(ctx context.Context, db database.IDB, orgID, collectionID string,
	grants []*entity.CollectionGrant) error {
	_, err := db.NewDelete().Model((*entity.CollectionGrant)(nil)).
		Where("org_id = ?", orgID).Where("collection_id = ?", collectionID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(grants) == 0 {
		return nil
	}
	if _, err = db.NewInsert().Model(&grants).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) RemoveUserGrants(ctx context.Context, db database.IDB, orgID, userID string) error {
	_, err := db.NewDelete().Model((*entity.CollectionGrant)(nil)).
		Where("org_id = ?", orgID).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// entitlementsQuery puts each way of having a permission in one list, then keeps the most for each person.
const entitlementsQuery = `
WITH granted AS (
    SELECT g.collection_id, g.user_id, g.permission
    FROM collection_grants AS g
    WHERE g.org_id = ?0 AND g.user_id IS NOT NULL
    UNION ALL
    SELECT g.collection_id, gm.user_id, g.permission
    FROM collection_grants AS g
    JOIN org_group_members AS gm ON gm.group_id = g.group_id
    WHERE g.org_id = ?0 AND g.group_id IS NOT NULL
    UNION ALL
    SELECT c.id, m.user_id, 'manage'
    FROM collections AS c
    JOIN org_members AS m ON m.org_id = c.org_id AND m.role = 'owner'
    WHERE c.org_id = ?0
)
SELECT r.collection_id, r.user_id, u.username, u.full_name, m.role,
    m.status AS member_status, u.status AS account_status,
    (ARRAY['connect', 'edit', 'manage'])[max(
        CASE r.permission WHEN 'connect' THEN 1 WHEN 'edit' THEN 2 ELSE 3 END)] AS permission
FROM granted AS r
JOIN collections AS c ON c.id = r.collection_id AND c.deleted_at IS NULL
JOIN org_members AS m ON m.org_id = ?0 AND m.user_id = r.user_id AND m.status <> 'removed'
JOIN users AS u ON u.id = r.user_id AND u.deleted_at IS NULL
WHERE (?1 = '' OR r.collection_id = ?1) AND (?2 = '' OR r.user_id = ?2)
GROUP BY r.collection_id, r.user_id, u.username, u.full_name, m.role, m.status, u.status
ORDER BY r.collection_id, lower(u.username)`

func (repo *collectionRepo) Entitlements(ctx context.Context, db database.IDB, orgID, collectionID, userID string) (
	[]*entity.CollectionEntitlement, error) {
	entitlements := []*entity.CollectionEntitlement{}
	if err := db.NewRaw(entitlementsQuery, orgID, collectionID, userID).Scan(ctx, &entitlements); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return entitlements, nil
}

func (repo *collectionRepo) ListWraps(ctx context.Context, db database.IDB, collectionID, userID string) (
	[]*entity.CollectionKeyWrap, error) {
	wraps := []*entity.CollectionKeyWrap{}
	err := db.NewSelect().Model(&wraps).
		ColumnExpr("collection_key_wrap.*, COALESCE(u.username, '') AS wrapper_username").
		Join("LEFT JOIN users AS u ON u.id = collection_key_wrap.wrapped_by").
		Where("collection_key_wrap.collection_id = ?", collectionID).
		Where("collection_key_wrap.user_id = ?", userID).
		OrderExpr("collection_key_wrap.key_version DESC").
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return wraps, nil
}

func (repo *collectionRepo) InsertWraps(ctx context.Context, db database.IDB, wraps []*entity.CollectionKeyWrap) error {
	if len(wraps) == 0 {
		return nil
	}
	if _, err := db.NewInsert().Model(&wraps).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *collectionRepo) KeyHolders(ctx context.Context, db database.IDB, orgID, collectionID string) (
	[]*entity.CollectionKeyHolder, error) {
	holders := []*entity.CollectionKeyHolder{}
	query := db.NewSelect().TableExpr("collection_key_wraps AS w").
		Join("JOIN collections AS c ON c.id = w.collection_id AND w.key_version = c.key_version").
		ColumnExpr("w.collection_id, w.user_id").
		Where("c.org_id = ?", orgID).Where("c.deleted_at IS NULL")
	if collectionID != "" {
		query = query.Where("c.id = ?", collectionID)
	}
	if err := query.Scan(ctx, &holders); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return holders, nil
}

func (repo *collectionRepo) DeleteUserWraps(ctx context.Context, db database.IDB, userID string) (int, error) {
	result, err := db.NewDelete().Model((*entity.CollectionKeyWrap)(nil)).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int(n), nil
}
