package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

type SettingRepo interface {
	GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		id string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetByKind(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		kind string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetByName(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		name string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetSingle(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)

	List(ctx context.Context, db database.IDB, scope *entity.ObjectScope, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, scope *entity.ObjectScope, ids []string, requireActive bool,
		opts ...bunex.SelectQueryOption) ([]*entity.Setting, error)

	// EnsureUnique makes sure there is at most one active setting in the given scope.
	// Inherited/imported settings are not taken into account.
	EnsureUnique(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		opts ...bunex.SelectQueryOption) error

	Insert(ctx context.Context, db database.IDB, setting *entity.Setting,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, setting *entity.Setting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, setting *entity.Setting,
		opts ...bunex.UpdateQueryOption) error
	UpdateClearDefaultFlag(ctx context.Context, db database.IDB, scope *entity.ObjectScope, typ base.SettingType,
		kind *string, exceptID string, opts ...bunex.UpdateQueryOption) error

	DeleteAllByObjects(ctx context.Context, db database.IDB, scope base.ObjectScopeType, objectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type settingRepo struct {
}

func NewSettingRepo() SettingRepo {
	return &settingRepo{}
}

func (repo *settingRepo) GetByID(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, id string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	opts = repo.applyFilter(opts, typ, id, requireActive)
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) GetByKind(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, kind string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if kind == "" {
		return nil, nil
	}
	opts = repo.applyFilter(opts, typ, "", requireActive)
	opts = repo.applyNameAndKindFilter(opts, "", kind)
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) GetByName(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, name string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if name == "" {
		return nil, nil
	}
	opts = repo.applyFilter(opts, typ, "", requireActive)
	opts = repo.applyNameAndKindFilter(opts, name, "")
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) get(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	theOpts := opts
	if scope != nil {
		switch scope.ScopeType {
		case base.ObjectScopeUser:
			theOpts = repo.applyDirectUserFilter(theOpts, scope)
		case base.ObjectScopeGlobal:
			theOpts = repo.applyGlobalFilter(theOpts)
		}
	}

	setting := &entity.Setting{}
	query := db.NewSelect().Model(setting)
	query = bunex.ApplySelect(query, theOpts...)

	err := query.Scan(ctx)
	if setting == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("Setting").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if hasChange, _ := repo.updateExpiredSetting(ctx, db, setting); hasChange {
		return repo.get(ctx, db, scope, opts...)
	}
	return setting, nil
}

func (repo *settingRepo) GetSingle(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	opts = repo.applyFilter(opts, typ, "", requireActive)

	// Global scope is the special case
	if scope != nil && scope.IsGlobalScope() {
		opts = repo.applyGlobalFilter(opts)
		setting, err := repo.get(ctx, db, scope, opts...)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return setting, nil
	}

	// For other scopes, we need to list satisfied settings upto the global scope,
	// then return the first matching one in the order of the scope upto global.
	settings, _, err := repo.List(ctx, db, scope, nil, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, hperrors.NewNotFound("Setting")
	}
	if len(settings) == 1 || scope == nil {
		return settings[0], nil
	}

	var globalSetting *entity.Setting
	for _, setting := range settings {
		switch scope.ScopeType {
		case base.ObjectScopeUser:
			if setting.ObjectID == scope.UserID { // user's direct setting has the highest priority
				return setting, nil
			}
		case base.ObjectScopeGlobal:
			return setting, nil
		}
		if globalSetting == nil && setting.ObjectID == "" {
			globalSetting = setting
			continue
		}
	}
	setting := gofn.Coalesce(globalSetting)
	if setting != nil {
		return setting, nil
	}
	return nil, hperrors.NewNotFound("Setting")
}

func (repo *settingRepo) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	theOpts := opts
	if scope != nil {
		switch scope.ScopeType {
		case base.ObjectScopeUser:
			theOpts = repo.applyDirectUserFilter(theOpts, scope)
		case base.ObjectScopeGlobal:
			theOpts = repo.applyGlobalFilter(theOpts)
		}
	}

	var settings []*entity.Setting
	query := db.NewSelect().Model(&settings)
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

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}

	if hasChange, _ := repo.updateExpiredSettings(ctx, db, settings); hasChange {
		return repo.List(ctx, db, scope, paging, opts...)
	}
	return settings, pagingMeta, nil
}

func (repo *settingRepo) ListByIDs(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	ids []string, requireActive bool, opts ...bunex.SelectQueryOption) ([]*entity.Setting, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	opts = append(opts, bunex.SelectWhere("setting.id IN (?)", bun.List(ids)))
	opts = repo.applyFilter(opts, "", "", requireActive)
	settings, _, err := repo.List(ctx, db, scope, nil, opts...)
	return settings, hperrors.Wrap(err)
}

func (repo *settingRepo) applyFilter(opts []bunex.SelectQueryOption, typ base.SettingType, id string,
	requireActive bool) []bunex.SelectQueryOption {
	if typ != "" {
		opts = append(opts, bunex.SelectWhere("setting.type = ?", typ))
	}
	if id != "" {
		opts = append(opts, bunex.SelectWhere("setting.id = ?", id))
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("setting.status = ?", base.SettingStatusActive))
	}
	return opts
}

func (repo *settingRepo) applyNameAndKindFilter(opts []bunex.SelectQueryOption,
	name, kind string) []bunex.SelectQueryOption {
	if name != "" {
		opts = append(opts, bunex.SelectWhere("LOWER(setting.name) = ?", strings.ToLower(name)))
	}
	if kind != "" {
		opts = append(opts, bunex.SelectWhere("setting.kind = ?", kind))
	}
	return opts
}

// applyDirectUserFilter filters settings belong to the user
func (repo *settingRepo) applyDirectUserFilter(opts []bunex.SelectQueryOption,
	scope *entity.ObjectScope) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("setting.object_id = ?", scope.UserID))
	return opts
}

// applyGlobalFilter filters settings belong to global scope
func (repo *settingRepo) applyGlobalFilter(opts []bunex.SelectQueryOption) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("(setting.scope = '' AND setting.object_id IS NULL)"))
	return opts
}

func (repo *settingRepo) EnsureUnique(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, opts ...bunex.SelectQueryOption) error {
	query := db.NewSelect().Model((*entity.Setting)(nil)).
		Where("setting.type = ?", typ).
		Where("setting.status = ?", base.SettingStatusActive)
	if scope != nil {
		scopeObjectID := scope.ScopeObjectID()
		if scopeObjectID != "" {
			query = query.Where("setting.object_id = ?", scopeObjectID)
		} else {
			query = query.Where("setting.object_id IS NULL")
		}
	}
	query = bunex.ApplySelect(query, opts...)

	count, err := query.Count(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return hperrors.Wrap(err)
	}
	if count > 1 {
		return hperrors.NewAlreadyExist("Setting")
	}
	return nil
}

func (repo *settingRepo) Insert(ctx context.Context, db database.IDB, setting *entity.Setting,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.Setting{setting}, opts...)
}

func (repo *settingRepo) InsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
	opts ...bunex.InsertQueryOption) error {
	if len(settings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&settings)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update res links for the settings
	err = repo.updateSettingResLinks(ctx, db, settings)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (repo *settingRepo) Upsert(ctx context.Context, db database.IDB, setting *entity.Setting,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Setting{setting}, conflictCols, updateCols, opts...)
}

func (repo *settingRepo) UpsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(settings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&settings)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update res links for the settings
	err = repo.updateSettingResLinks(ctx, db, settings)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (repo *settingRepo) Update(ctx context.Context, db database.IDB, setting *entity.Setting,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(setting).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update res links for the setting
	err = repo.updateSettingResLinks(ctx, db, []*entity.Setting{setting})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (repo *settingRepo) UpdateClearDefaultFlag(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, kind *string, exceptID string, opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model((*entity.Setting)(nil)).
		Where("setting.type = ?", typ).
		Where("setting.is_default = true").
		Set("is_default = false")
	if kind != nil {
		if *kind == "" {
			query = query.Where("(setting.kind IS NULL OR setting.kind = '')")
		} else {
			query = query.Where("setting.kind = ?", *kind)
		}
	}
	if exceptID != "" {
		query = query.Where("setting.id != ?", exceptID)
	}
	if scope != nil {
		scopeObjectID := scope.ScopeObjectID()
		if scopeObjectID != "" {
			query = query.Where("setting.object_id = ?", scopeObjectID)
		} else {
			query = query.Where("setting.object_id IS NULL")
		}
	}
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) updateExpiredSetting(ctx context.Context, db database.IDB,
	setting *entity.Setting) (bool, error) {
	if setting == nil {
		return false, nil
	}
	return repo.updateExpiredSettings(ctx, db, []*entity.Setting{setting})
}

func (repo *settingRepo) updateExpiredSettings(ctx context.Context, db database.IDB,
	settings []*entity.Setting) (hasChange bool, err error) {
	for _, setting := range settings {
		if setting.IsStatusDirty() {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return false, nil
	}
	query := db.NewUpdate().Model((*entity.Setting)(nil)).
		Set("status = ?", base.SettingStatusExpired).
		Where("status = ? AND expire_at < NOW()", base.SettingStatusActive)

	_, err = query.Exec(ctx)
	if err != nil {
		return hasChange, hperrors.Wrap(err)
	}
	return hasChange, nil
}

func (repo *settingRepo) updateSettingResLinks(ctx context.Context, db database.IDB, settings []*entity.Setting) error {
	if len(settings) == 0 {
		return nil
	}
	settingIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		settingIDs = append(settingIDs, setting.ID)
	}

	newLinks := make([]*entity.ResLink, 0, len(settings)*2) //nolint:mnd
	for _, setting := range settings {
		links, err := setting.GetResourceLinks()
		if err != nil {
			return hperrors.Wrap(err)
		}
		newLinks = append(newLinks, links...)
	}

	if len(newLinks) == 0 { // delete all current links
		query := db.NewDelete().Model((*entity.ResLink)(nil)).
			Where("res_link.src_type = ?", base.ResourceTypeSetting).
			Where("res_link.src_id IN (?)", bun.List(settingIDs))

		_, err := query.Exec(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	}

	var currLinks []*entity.ResLink
	selQuery := db.NewSelect().Model(&currLinks).
		Where("res_link.src_type = ?", base.ResourceTypeSetting).
		Where("res_link.src_id IN (?)", bun.List(settingIDs))
	err := selQuery.Scan(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	upsertingLinks := resLinksToUpsert(newLinks, currLinks, timeutil.NowUTC())
	if len(upsertingLinks) > 0 {
		upsertQuery := db.NewInsert().Model(&upsertingLinks)
		upsertQuery = bunex.ApplyUpsert(upsertQuery, entity.ResLinkUpsertingConflictCols,
			entity.ResLinkUpsertingUpdateCols)

		_, err = upsertQuery.Exec(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

// resLinksToUpsert is what brings the settings' links from what they are to
// what they should be: each wanted link that is new or changed, and each
// current one no longer wanted, deleted. A link wanted twice - one target told
// of success and of failure - is written once: one statement may not upsert a
// row twice.
func resLinksToUpsert(wanted, current []*entity.ResLink, now time.Time) []*entity.ResLink {
	currentByKey := make(map[string]*entity.ResLink, len(current))
	for _, link := range current {
		currentByKey[link.GetKey()] = link
	}

	upserting := make([]*entity.ResLink, 0, len(wanted))
	seen := make(map[string]struct{}, len(wanted))
	for _, link := range wanted {
		key := link.GetKey()
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		if curr, ok := currentByKey[key]; ok {
			delete(currentByKey, key)
			if curr.Index != link.Index || curr.Data != link.Data {
				upserting = append(upserting, link)
			}
		} else {
			upserting = append(upserting, link)
		}
	}

	for _, link := range currentByKey {
		link.DeletedAt = now
		upserting = append(upserting, link)
	}
	return upserting
}

func (repo *settingRepo) DeleteAllByObjects(ctx context.Context, db database.IDB,
	scope base.ObjectScopeType, objectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(objectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Setting)(nil)).
		Where("setting.scope = ?", scope).
		Where("setting.object_id IN (?)", bun.List(objectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Setting)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
