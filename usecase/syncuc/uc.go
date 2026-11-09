package syncuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/usecase/syncuc/syncdto"
)

const (
	changesDefaultLimit = 1000
	changesMaxLimit     = 5000
)

type UC struct {
	db             *database.DB
	syncService    syncservice.Service
	orgRepo        repository.OrganizationRepo
	collectionRepo repository.CollectionRepo
}

func New(db *database.DB, syncService syncservice.Service, orgRepo repository.OrganizationRepo,
	collectionRepo repository.CollectionRepo) *UC {
	return &UC{db: db, syncService: syncService, orgRepo: orgRepo, collectionRepo: collectionRepo}
}

// CheckAccess lets a person read a space: their personal one, or a collection served to them.
func (uc *UC) CheckAccess(ctx context.Context, auth *basedto.Auth, spaceID string) error {
	return uc.checkAccess(ctx, auth, spaceID, false)
}

// checkAccess lets a person reach a space of theirs, or a collection they are an active member with a grant of: to
// write, with Edit or Manage. Any other space is simply not found: saying more would tell a caller which ids exist.
func (uc *UC) checkAccess(ctx context.Context, auth *basedto.Auth, spaceID string, write bool) error {
	space, err := uc.syncService.GetSpace(ctx, spaceID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	switch space.Type {
	case base.SpaceTypePersonal:
		if space.OwnerUserID == auth.UserID() {
			return nil
		}
	case base.SpaceTypeCollection:
		permission, err := uc.collectionPermission(ctx, auth, space)
		if err != nil {
			return err
		}
		if write && permission != "" && !permission.AtLeast(base.CollectionPermissionEdit) {
			return hperrors.Wrap(hperrors.ErrCollectionReadOnly)
		}
		if permission != "" {
			return nil
		}
	}
	return hperrors.NewNotFound("Space")
}

// collectionPermission is what the person may do in the collection now: nothing unless the organization is active and
// they are served.
func (uc *UC) collectionPermission(ctx context.Context, auth *basedto.Auth, space *entity.Space) (
	base.CollectionPermission, error) {
	org, err := uc.orgRepo.GetByID(ctx, uc.db, space.OrgID)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if org == nil || org.Status != base.OrgStatusActive {
		return "", nil
	}
	entitlements, err := uc.collectionRepo.Entitlements(ctx, uc.db, space.OrgID, space.ID, auth.UserID())
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if len(entitlements) == 0 || !entitlements[0].Served() {
		return "", nil
	}
	return entitlements[0].Permission, nil
}

// ListSpaces returns the spaces of the person: their personal one, made the first time they ask, then the collections
// of each organization served to them.
func (uc *UC) ListSpaces(ctx context.Context, auth *basedto.Auth) (*syncdto.ListSpacesResp, error) {
	space, err := uc.syncService.EnsurePersonalSpace(ctx, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	totals, err := uc.syncService.GetTotals(ctx, space.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := []*syncdto.SpaceResp{syncdto.TransformSpace(space, totals)}
	memberships, err := uc.orgRepo.ListMembershipsOfUser(ctx, uc.db, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, membership := range memberships {
		collections, err := uc.servedCollections(ctx, auth, membership.OrgID)
		if err != nil {
			return nil, err
		}
		data = append(data, collections...)
	}
	return &syncdto.ListSpacesResp{Data: data}, nil
}

func (uc *UC) servedCollections(ctx context.Context, auth *basedto.Auth, orgID string) ([]*syncdto.SpaceResp, error) {
	entitlements, err := uc.collectionRepo.Entitlements(ctx, uc.db, orgID, "", auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(entitlements) == 0 {
		return nil, nil
	}
	collections, err := uc.collectionRepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Collection, len(collections))
	for _, collection := range collections {
		byID[collection.ID] = collection
	}
	stale, err := uc.needingNewKey(ctx, orgID, entitlements)
	if err != nil {
		return nil, err
	}
	var out []*syncdto.SpaceResp
	for _, e := range entitlements {
		collection := byID[e.CollectionID]
		if collection == nil || !e.Served() {
			continue
		}
		totals, err := uc.syncService.GetTotals(ctx, collection.ID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp := syncdto.TransformCollection(collection, e.Permission, totals)
		resp.NeedsNewKey = e.Permission == base.CollectionPermissionManage && stale[collection.ID]
		out = append(out, resp)
	}
	return out, nil
}

// needingNewKey gives the organization's collections whose key must be replaced, of those the person manages
// (mine): none, without asking, if they manage none.
func (uc *UC) needingNewKey(ctx context.Context, orgID string, mine []*entity.CollectionEntitlement) (
	map[string]bool, error) {
	manages := false
	for _, e := range mine {
		manages = manages || e.Permission == base.CollectionPermissionManage
	}
	if !manages {
		return nil, nil
	}
	entitlements, err := uc.collectionRepo.Entitlements(ctx, uc.db, orgID, "", "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	holders, err := uc.collectionRepo.KeyHolders(ctx, uc.db, orgID, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entitled := map[string]map[string]*entity.CollectionEntitlement{}
	for _, e := range entitlements {
		if entitled[e.CollectionID] == nil {
			entitled[e.CollectionID] = map[string]*entity.CollectionEntitlement{}
		}
		entitled[e.CollectionID][e.UserID] = e
	}
	held := map[string]map[string]bool{}
	for _, h := range holders {
		if held[h.CollectionID] == nil {
			held[h.CollectionID] = map[string]bool{}
		}
		held[h.CollectionID][h.UserID] = true
	}
	out := map[string]bool{}
	for id, holders := range held {
		out[id] = entity.NeedsNewKey(entitled[id], holders)
	}
	return out, nil
}

func (uc *UC) ListChanges(ctx context.Context, auth *basedto.Auth, spaceID string, since int64,
	limit int) (*syncdto.ListChangesResp, error) {
	if err := uc.CheckAccess(ctx, auth, spaceID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = changesDefaultLimit
	}
	limit = min(limit, changesMaxLimit)
	changes, err := uc.syncService.ChangesSince(ctx, spaceID, since, limit)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	objects := make([]*syncdto.ObjectResp, 0, len(changes.Objects))
	for _, object := range changes.Objects {
		objects = append(objects, &syncdto.ObjectResp{
			Name:    object.Name,
			ETag:    object.ETag,
			Size:    object.Size,
			Deleted: !object.IsLive(),
		})
	}
	return &syncdto.ListChangesResp{Data: &syncdto.ChangesResp{
		Revision: changes.Revision,
		More:     changes.More,
		Objects:  objects,
	}}, nil
}

// GetObject returns the file, or only its line when the caller holds that very version: etag is the one it holds.
func (uc *UC) GetObject(ctx context.Context, auth *basedto.Auth, spaceID, name,
	etag string) (*entity.SyncObject, []byte, error) {
	if err := uc.CheckAccess(ctx, auth, spaceID); err != nil {
		return nil, nil, err
	}
	if etag != "" {
		object, err := uc.syncService.Stat(ctx, spaceID, name)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		if object.ETag == etag {
			return object, nil, nil
		}
	}
	object, data, err := uc.syncService.Read(ctx, spaceID, name)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck // wrapped by the service, or a plain internal error
	}
	return object, data, nil
}

func (uc *UC) PutObject(ctx context.Context, auth *basedto.Auth, spaceID, name string, data []byte,
	cond syncservice.Condition) (*syncdto.PutObjectResp, error) {
	if err := uc.checkAccess(ctx, auth, spaceID, true); err != nil {
		return nil, err
	}
	object, err := uc.syncService.Write(ctx, spaceID, name, data, cond)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &syncdto.PutObjectResp{Data: &syncdto.PutObjectData{ETag: object.ETag, Revision: object.Revision}}, nil
}

func (uc *UC) DeleteObject(ctx context.Context, auth *basedto.Auth, spaceID, name string,
	cond syncservice.Condition) (*syncdto.DeleteObjectResp, error) {
	if err := uc.checkAccess(ctx, auth, spaceID, true); err != nil {
		return nil, err
	}
	if err := uc.syncService.Remove(ctx, spaceID, name, cond); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &syncdto.DeleteObjectResp{}, nil
}
