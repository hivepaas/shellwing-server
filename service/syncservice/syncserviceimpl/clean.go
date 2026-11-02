package syncserviceimpl

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/blobstore"
	"github.com/hivepaas/shellwing-server/pkg/dblock"
	"github.com/hivepaas/shellwing-server/pkg/safego"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

const (
	cleanLockName = "sync-clean"
	trashBatch    = 500
	sweepBatch    = 500
	sweepInterval = 24 * time.Hour
	// Content this young may belong to a write whose line is about to be saved.
	sweepMinAge     = 24 * time.Hour
	blobsPrefix     = "spaces/"
	blobKeySegments = 4 // spaces, the space's id, at least one of the name, the ETag
)

// Start checks the store and begins cleaning it. A store that cannot be used stops the server's start.
func (s *service) Start(ctx context.Context) error {
	if err := s.store.Check(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	// Not the start's context, which ends with the start. Stop ends this one.
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancel = cancel
	s.wg.Add(1)
	safego.Go("syncClean", func() {
		defer s.wg.Done()
		s.cleanLoop(runCtx)
	})
	return nil
}

func (s *service) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	return nil
}

func (s *service) cleanLoop(ctx context.Context) {
	interval := config.Current().Sync.CleanInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	lastSweep := timeutil.NowUTC()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		now := timeutil.NowUTC()
		sweep := now.Sub(lastSweep) >= sweepInterval
		if err := s.cleanOnce(ctx, sweep); err != nil {
			s.logger.Errorf("sync: failed to clean the store: %v", err)
			continue
		}
		if sweep {
			lastSweep = now
		}
	}
}

// cleanOnce runs on one instance at a time: another holding the lock is doing the same work.
func (s *service) cleanOnce(ctx context.Context, sweep bool) error {
	lock, acquired, err := dblock.TryAcquire(ctx, s.db, cleanLockName)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !acquired {
		return nil
	}
	defer func() { _ = lock.Release(context.WithoutCancel(ctx)) }()

	if err = s.emptyTrash(ctx); err != nil {
		return err
	}
	if sweep {
		return s.sweep(ctx)
	}
	return nil
}

// emptyTrash removes the content whose time has come, unless a line names it: the same content can be written back.
func (s *service) emptyTrash(ctx context.Context) error {
	for {
		items, err := s.syncRepo.ListDueTrash(ctx, s.db, timeutil.NowUTC(), trashBatch)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, item := range items {
			if err = s.emptyTrashItem(ctx, item); err != nil {
				return err
			}
		}
		if len(items) < trashBatch {
			return nil
		}
	}
}

func (s *service) emptyTrashItem(ctx context.Context, item *entity.SyncTrash) error {
	current, err := s.syncRepo.GetObject(ctx, s.db, item.SpaceID, item.Name)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !current.IsLive() || current.ETag != item.ETag {
		err = s.store.Delete(ctx, entity.SyncBlobKey(item.SpaceID, item.Name, item.ETag))
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return hperrors.Wrap(s.syncRepo.DeleteTrash(ctx, s.db, item))
}

// splitBlobKey undoes entity.SyncBlobKey. A key that is not one of its keys gives false.
func splitBlobKey(key string) (spaceID, name, etag string, ok bool) {
	segments := strings.Split(key, "/")
	if len(segments) < blobKeySegments || segments[0] != strings.TrimSuffix(blobsPrefix, "/") {
		return "", "", "", false
	}
	last := len(segments) - 1
	return segments[1], strings.Join(segments[2:last], "/"), segments[last], true
}

type sweepEntry struct {
	key, name, etag string
}

// sweep removes what no line names and the trash has not noted: what a crash left between storing content and saving
// its line. It walks the store, so it is rare.
func (s *service) sweep(ctx context.Context) error {
	var spaceID string
	var batch []sweepEntry
	flush := func() error {
		err := s.sweepBatch(ctx, spaceID, batch)
		batch = batch[:0]
		return err
	}
	cutoff := timeutil.NowUTC().Add(-sweepMinAge)
	err := s.store.List(ctx, blobsPrefix, func(blob blobstore.Blob) error {
		space, name, etag, ok := splitBlobKey(blob.Key)
		if !ok || blob.Modified.After(cutoff) {
			return nil
		}
		if space != spaceID || len(batch) >= sweepBatch {
			if err := flush(); err != nil {
				return err
			}
			spaceID = space
		}
		batch = append(batch, sweepEntry{key: blob.Key, name: name, etag: etag})
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return flush()
}

func (s *service) sweepBatch(ctx context.Context, spaceID string, batch []sweepEntry) error {
	if len(batch) == 0 {
		return nil
	}
	names := make([]string, 0, len(batch))
	for i := range batch {
		names = append(names, batch[i].name)
	}
	current, err := s.syncRepo.CurrentETags(ctx, s.db, spaceID, names)
	if err != nil {
		return hperrors.Wrap(err)
	}
	trashed, err := s.syncRepo.TrashedETags(ctx, s.db, spaceID, names)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for i := range batch {
		entry := &batch[i]
		if current[entry.name] == entry.etag || slices.Contains(trashed[entry.name], entry.etag) {
			continue
		}
		if err = s.store.Delete(ctx, entry.key); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
