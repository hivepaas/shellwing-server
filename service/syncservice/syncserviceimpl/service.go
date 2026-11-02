package syncserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/blobstore"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
)

// NewStore makes the store the configuration names.
func NewStore(cfg *config.Config) (blobstore.Store, error) {
	if cfg.Sync.Store.Type == config.SyncStoreS3 {
		store := &cfg.Sync.Store
		return blobstore.NewS3(&blobstore.S3Config{ //nolint:wrapcheck // the package's own
			Endpoint:  store.Endpoint,
			Region:    store.Region,
			Bucket:    store.Bucket,
			Prefix:    store.Prefix,
			AccessKey: store.AccessKey,
			SecretKey: store.SecretKey,
			PathStyle: store.PathStyle,
		})
	}
	return blobstore.NewLocal(cfg.SyncStorePath()) //nolint:wrapcheck // the package's own
}

func New(
	db *database.DB,
	store blobstore.Store,
	spaceRepo repository.SpaceRepo,
	syncRepo repository.SyncRepo,
	eventBus systemeventbusservice.Service,
	logger logging.Logger,
) syncservice.Service {
	return &service{
		db:        db,
		store:     store,
		spaceRepo: spaceRepo,
		syncRepo:  syncRepo,
		eventBus:  eventBus,
		logger:    logger,
	}
}

type service struct {
	db        *database.DB
	store     blobstore.Store
	spaceRepo repository.SpaceRepo
	syncRepo  repository.SyncRepo
	eventBus  systemeventbusservice.Service
	logger    logging.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (s *service) EnsurePersonalSpace(ctx context.Context, userID string) (*entity.Space, error) {
	space, err := s.spaceRepo.GetPersonal(ctx, s.db, userID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if space != nil {
		return space, nil
	}
	err = transaction.Execute(ctx, s.db, func(tx database.Tx) error {
		timeNow := timeutil.NowUTC()
		err := s.spaceRepo.InsertIfAbsent(ctx, tx, &entity.Space{
			ID:          gofn.Must(ulid.NewStringULID()),
			Type:        base.SpaceTypePersonal,
			OwnerUserID: userID,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		space, err = s.spaceRepo.GetPersonal(ctx, tx, userID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if space == nil {
			return hperrors.NewNotFound("Space")
		}
		return s.syncRepo.EnsureSpace(ctx, tx, space.ID)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return space, nil
}

func (s *service) GetSpace(ctx context.Context, spaceID string) (*entity.Space, error) {
	space, err := s.spaceRepo.GetByID(ctx, s.db, spaceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return space, nil
}

func (s *service) GetTotals(ctx context.Context, spaceID string) (*entity.SyncSpace, error) {
	totals, err := s.syncRepo.GetSpace(ctx, s.db, spaceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return totals, nil
}

func checkCondition(object *entity.SyncObject, cond syncservice.Condition) error {
	switch {
	case cond.IfAbsent && object.IsLive():
		return hperrors.Wrap(hperrors.ErrSyncObjectChanged)
	case cond.IfMatch != "" && (!object.IsLive() || object.ETag != cond.IfMatch):
		return hperrors.Wrap(hperrors.ErrSyncObjectChanged)
	}
	return nil
}

func (s *service) Write(ctx context.Context, spaceID, name string, data []byte,
	cond syncservice.Condition) (*entity.SyncObject, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	cfg := config.Current()
	if int64(len(data)) > cfg.Sync.MaxObjectBytes {
		return nil, hperrors.Wrap(hperrors.ErrSyncObjectTooLarge)
	}
	object := &entity.SyncObject{SpaceID: spaceID, Name: name, ETag: etagOf(data), Size: int64(len(data))}
	if err := describe(object, data); err != nil {
		return nil, err
	}

	// A write that is going to lose need not cost an upload. The transaction below is what decides.
	current, err := s.syncRepo.GetObject(ctx, s.db, spaceID, name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = checkCondition(current, cond); err != nil {
		return nil, err
	}

	// The content goes first, under a key of its own, so that no line ever names what is not there. Nobody reads this
	// key until the line names it. A crash here leaves content no line names, which the sweep removes.
	key := entity.SyncBlobKey(spaceID, name, object.ETag)
	if err = s.store.Put(ctx, key, data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	changed := false
	err = transaction.Execute(ctx, s.db, func(tx database.Tx) error {
		var err error
		changed, err = s.writeLine(ctx, tx, object, cond, cfg)
		return err
	})
	if err != nil {
		// Maybe this very content is what a line names by now, so it only goes to the trash, which checks.
		s.noteTrash(ctx, spaceID, name, object.ETag)
		return nil, hperrors.Wrap(err)
	}
	if changed {
		s.publish(ctx, spaceID, object.Revision)
	}
	return object, nil
}

// writeLine decides the write. object takes the revision and times of the file as it stands afterwards.
func (s *service) writeLine(ctx context.Context, tx database.IDB, object *entity.SyncObject,
	cond syncservice.Condition, cfg *config.Config) (bool, error) {
	totals, err := s.syncRepo.LockSpace(ctx, tx, object.SpaceID)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	current, err := s.syncRepo.GetObject(ctx, tx, object.SpaceID, object.Name)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if err = checkCondition(current, cond); err != nil {
		return false, err
	}
	timeNow := timeutil.NowUTC()
	object.CreatedAt, object.UpdatedAt = timeNow, timeNow
	if current != nil {
		object.CreatedAt = current.CreatedAt
	}
	if current.IsLive() && current.ETag == object.ETag {
		object.Revision, object.UpdatedAt = current.Revision, current.UpdatedAt
		return false, nil
	}

	counts(totals, current, -1)
	counts(totals, object, 1)
	if limit := cfg.Sync.MaxSpaceBytes; limit > 0 && totals.BytesUsed > limit {
		// A write that frees room is always let through, so a full space can be brought back under.
		if current == nil || !current.IsLive() || object.Size > current.Size {
			return false, hperrors.Wrap(hperrors.ErrSyncSpaceFull)
		}
	}
	totals.Revision++
	totals.UpdatedAt = timeNow
	object.Revision = totals.Revision
	if err = s.syncRepo.UpsertObject(ctx, tx, object); err != nil {
		return false, hperrors.Wrap(err)
	}
	if err = s.syncRepo.UpdateSpace(ctx, tx, totals); err != nil {
		return false, hperrors.Wrap(err)
	}
	if current.IsLive() {
		err = s.syncRepo.AddTrash(ctx, tx, object.SpaceID, object.Name, current.ETag,
			timeNow.Add(cfg.Sync.TrashDelay))
		if err != nil {
			return false, hperrors.Wrap(err)
		}
	}
	return true, nil
}

func (s *service) noteTrash(ctx context.Context, spaceID, name, etag string) {
	deleteAfter := timeutil.NowUTC().Add(config.Current().Sync.TrashDelay)
	if err := s.syncRepo.AddTrash(ctx, s.db, spaceID, name, etag, deleteAfter); err != nil {
		// The sweep finds it instead.
		s.logger.Errorf("sync: failed to note content for removal: %v", err)
	}
}

func (s *service) Stat(ctx context.Context, spaceID, name string) (*entity.SyncObject, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	object, err := s.syncRepo.GetObject(ctx, s.db, spaceID, name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !object.IsLive() {
		return nil, hperrors.NewNotFound("File")
	}
	return object, nil
}

func (s *service) Read(ctx context.Context, spaceID, name string) (*entity.SyncObject, []byte, error) {
	object, err := s.Stat(ctx, spaceID, name)
	if err != nil {
		return nil, nil, err
	}
	data, err := s.store.Get(ctx, entity.SyncBlobKey(spaceID, name, object.ETag))
	if err != nil {
		if errors.Is(err, blobstore.ErrNotFound) {
			return nil, nil, fmt.Errorf("the store has lost %s of space %s: %w", name, spaceID, err)
		}
		return nil, nil, hperrors.Wrap(err)
	}
	return object, data, nil
}

func (s *service) Remove(ctx context.Context, spaceID, name string, cond syncservice.Condition) error {
	if err := validateName(name); err != nil {
		return err
	}
	cfg := config.Current()
	var revision int64
	err := transaction.Execute(ctx, s.db, func(tx database.Tx) error {
		revision = 0
		totals, err := s.syncRepo.LockSpace(ctx, tx, spaceID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		current, err := s.syncRepo.GetObject(ctx, tx, spaceID, name)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = checkCondition(current, cond); err != nil {
			return err
		}
		if !current.IsLive() {
			return nil
		}
		timeNow := timeutil.NowUTC()
		counts(totals, current, -1)
		totals.Revision++
		totals.UpdatedAt = timeNow
		current.Revision = totals.Revision
		current.UpdatedAt = timeNow
		current.DeletedAt = timeNow
		if err = s.syncRepo.UpsertObject(ctx, tx, current); err != nil {
			return hperrors.Wrap(err)
		}
		if err = s.syncRepo.UpdateSpace(ctx, tx, totals); err != nil {
			return hperrors.Wrap(err)
		}
		revision = totals.Revision
		return s.syncRepo.AddTrash(ctx, tx, spaceID, name, current.ETag, timeNow.Add(cfg.Sync.TrashDelay))
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if revision > 0 {
		s.publish(ctx, spaceID, revision)
	}
	return nil
}

func (s *service) ChangesSince(ctx context.Context, spaceID string, since int64,
	limit int) (*syncservice.Changes, error) {
	totals, err := s.syncRepo.GetSpace(ctx, s.db, spaceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// One more than asked for tells whether there is more.
	objects, err := s.syncRepo.ListChanges(ctx, s.db, spaceID, since, limit+1)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	changes := &syncservice.Changes{Revision: totals.Revision, Objects: objects}
	if len(objects) > limit {
		changes.Objects = objects[:limit]
		changes.More = true
		changes.Revision = changes.Objects[limit-1].Revision
	}
	return changes, nil
}

// publish tells every instance. A notice that is lost costs a delay, not data: clients also ask on their own.
func (s *service) publish(ctx context.Context, spaceID string, revision int64) {
	err := s.eventBus.Publish(ctx, base.SystemEventSyncChanged, fmt.Sprintf("%s %d", spaceID, revision))
	if err != nil {
		s.logger.Errorf("sync: failed to publish a change: %v", err)
	}
}
