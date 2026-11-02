package syncuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/usecase/syncuc/syncdto"
)

const (
	changesDefaultLimit = 1000
	changesMaxLimit     = 5000
)

type UC struct {
	syncService syncservice.Service
}

func New(syncService syncservice.Service) *UC {
	return &UC{syncService: syncService}
}

// CheckAccess lets a person reach a space of theirs. Any other space is simply not found: saying more would tell a
// caller which ids exist. Collections come with the permissions that govern them.
func (uc *UC) CheckAccess(ctx context.Context, auth *basedto.Auth, spaceID string) error {
	space, err := uc.syncService.GetSpace(ctx, spaceID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if space.Type != base.SpaceTypePersonal || space.OwnerUserID != auth.UserID() {
		return hperrors.NewNotFound("Space")
	}
	return nil
}

// ListSpaces returns the spaces of the person: for now their personal one, made the first time they ask.
func (uc *UC) ListSpaces(ctx context.Context, auth *basedto.Auth) (*syncdto.ListSpacesResp, error) {
	space, err := uc.syncService.EnsurePersonalSpace(ctx, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	totals, err := uc.syncService.GetTotals(ctx, space.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &syncdto.ListSpacesResp{Data: []*syncdto.SpaceResp{syncdto.TransformSpace(space, totals)}}, nil
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
	if err := uc.CheckAccess(ctx, auth, spaceID); err != nil {
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
	if err := uc.CheckAccess(ctx, auth, spaceID); err != nil {
		return nil, err
	}
	if err := uc.syncService.Remove(ctx, spaceID, name, cond); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &syncdto.DeleteObjectResp{}, nil
}
