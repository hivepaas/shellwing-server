package syncserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/blobstore"
	"github.com/hivepaas/shellwing-server/pkg/dblock"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
)

// These need a real PostgreSQL with the migrations applied: what they check is that the row lock and the conditions
// hold under real concurrency. Point SHW_TEST_PG_DSN at one to enable them. Each test makes its own user, and leaves
// what it made behind in a test database.
type testEnv struct {
	s      *service
	store  blobstore.Store
	space  string
	events *publishedEvents
}

type publishedEvents struct {
	systemeventbusservice.Service
	mu       sync.Mutex
	payloads []string
}

func (p *publishedEvents) Publish(_ context.Context, _ base.SystemEventType, payload ...string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.payloads = append(p.payloads, payload...)
	return nil
}

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping sync test")
	}
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	bunDB := bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })
	return &database.DB{DB: bunDB}
}

func newEnv(t *testing.T) *testEnv {
	t.Helper()
	db := openTestDB(t)
	ctx := context.Background()
	config.SetCurrent(&config.Config{Sync: config.Sync{MaxObjectBytes: 1 << 20, TrashDelay: time.Hour}})
	store, err := blobstore.NewLocal(filepath.Join(t.TempDir(), "blobs"))
	require.NoError(t, err)
	require.NoError(t, store.Check(ctx))
	events := &publishedEvents{}
	s := New(db, store, repository.NewSpaceRepo(), repository.NewSyncRepo(), events, logging.GlobalLogger()).(*service)

	userID := gofn.Must(ulid.NewStringULID())
	_, err = db.NewInsert().Model(&entity.User{
		ID: userID, Username: userID, Role: base.UserRoleMember, Status: base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
	}).Exec(ctx)
	require.NoError(t, err)
	space, err := s.EnsurePersonalSpace(ctx, userID)
	require.NoError(t, err)
	return &testEnv{s: s, store: store, space: space.ID, events: events}
}

func recordBytes(t *testing.T, id, kind string, live bool, content string) []byte {
	t.Helper()
	record := map[string]any{"id": id, "kind": kind, "space": "personal", "key": 1, "hlc": "1"}
	if live {
		record["data"] = content
	}
	data, err := json.Marshal(record)
	require.NoError(t, err)
	return data
}

func (e *testEnv) totals(t *testing.T) *entity.SyncSpace {
	t.Helper()
	totals, err := e.s.GetTotals(context.Background(), e.space)
	require.NoError(t, err)
	return totals
}

func TestWriteReadRemoveAndTotals(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	host := recordBytes(t, "h1", "host", true, "sealed")
	object, err := e.s.Write(ctx, e.space, "records/h1.sw", host, syncservice.Condition{IfAbsent: true})
	require.NoError(t, err)
	assert.Equal(t, int64(1), object.Revision)
	_, err = e.s.Write(ctx, e.space, "records/o1.sw", recordBytes(t, "o1", "other", true, "x"), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "format.json", []byte("{}"), syncservice.Condition{})
	require.NoError(t, err)

	totals := e.totals(t)
	assert.Equal(t, int64(3), totals.Revision)
	assert.Equal(t, 2, totals.RecordCount, "format.json is no record")
	assert.Equal(t, 1, totals.HostCount)

	got, data, err := e.s.Read(ctx, e.space, "records/h1.sw")
	require.NoError(t, err)
	assert.Equal(t, host, data)
	assert.Equal(t, object.ETag, got.ETag)

	// The same content again changes nothing.
	again, err := e.s.Write(ctx, e.space, "records/h1.sw", host, syncservice.Condition{IfMatch: object.ETag})
	require.NoError(t, err)
	assert.Equal(t, object.Revision, again.Revision)
	assert.Equal(t, int64(3), e.totals(t).Revision)

	// A client's delete is a tombstone: the file stays, and is no longer counted as a host.
	tombstone, err := e.s.Write(ctx, e.space, "records/h1.sw", recordBytes(t, "h1", "host", false, ""),
		syncservice.Condition{IfMatch: object.ETag})
	require.NoError(t, err)
	assert.True(t, tombstone.Tombstone)
	totals = e.totals(t)
	assert.Equal(t, 0, totals.HostCount)
	assert.Equal(t, 1, totals.RecordCount)

	// Then the file goes.
	require.NoError(t, e.s.Remove(ctx, e.space, "records/h1.sw", syncservice.Condition{IfMatch: tombstone.ETag}))
	_, _, err = e.s.Read(ctx, e.space, "records/h1.sw")
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	require.NoError(t, e.s.Remove(ctx, e.space, "records/h1.sw", syncservice.Condition{}), "already gone is no error")
	require.ErrorIs(t, e.s.Remove(ctx, e.space, "records/h1.sw", syncservice.Condition{IfMatch: tombstone.ETag}),
		hperrors.ErrSyncObjectChanged)

	// And can come back, as a new file.
	_, err = e.s.Write(ctx, e.space, "records/h1.sw", host, syncservice.Condition{IfAbsent: true})
	require.NoError(t, err)
	assert.Equal(t, 1, e.totals(t).HostCount)
	assert.Len(t, e.events.payloads, int(e.totals(t).Revision), "one notice per revision")
}

func TestWriteRefusesWhatItShould(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	for _, name := range []string{"", "../x", "records/../x", "records/", "other/x", "records/a/b", "/format.json",
		"records/.hidden", "records/" + string(make([]byte, 300))} {
		_, err := e.s.Write(ctx, e.space, name, []byte("{}"), syncservice.Condition{})
		assert.ErrorIs(t, err, hperrors.ErrSyncObjectNameInvalid, name)
	}
	for name, data := range map[string][]byte{
		"records/a.sw": []byte("not json"),
		"records/b.sw": recordBytes(t, "another", "host", true, "x"),
		"records/c.sw": recordBytes(t, "c", "", true, "x"),
	} {
		_, err := e.s.Write(ctx, e.space, name, data, syncservice.Condition{})
		assert.ErrorIs(t, err, hperrors.ErrSyncRecordInvalid, name)
	}
	assert.Equal(t, int64(0), e.totals(t).Revision, "nothing refused counts")

	_, err := e.s.Write(ctx, e.space, "attachments/big.sw",
		make([]byte, config.Current().Sync.MaxObjectBytes+1), syncservice.Condition{})
	require.ErrorIs(t, err, hperrors.ErrSyncObjectTooLarge)

	_, err = e.s.Write(ctx, e.space, "format.json", []byte("{}"), syncservice.Condition{IfMatch: "nope"})
	require.ErrorIs(t, err, hperrors.ErrSyncObjectChanged)
}

// Of many devices writing over the same version, exactly one wins. The rest are told, and their content is not what
// the file holds.
func TestOnlyOneConditionalWriteWins(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	first, err := e.s.Write(ctx, e.space, "records/r.sw", recordBytes(t, "r", "host", true, "base"),
		syncservice.Condition{})
	require.NoError(t, err)

	const writers = 24
	var won, lost atomic.Int32
	var winner atomic.Value
	var wg sync.WaitGroup
	for i := range writers {
		wg.Go(func() {
			object, err := e.s.Write(ctx, e.space, "records/r.sw",
				recordBytes(t, "r", "host", true, fmt.Sprintf("device %d", i)),
				syncservice.Condition{IfMatch: first.ETag})
			if err != nil {
				assert.ErrorIs(t, err, hperrors.ErrSyncObjectChanged)
				lost.Add(1)
				return
			}
			won.Add(1)
			winner.Store(object.ETag)
		})
	}
	wg.Wait()
	assert.Equal(t, int32(1), won.Load())
	assert.Equal(t, int32(writers-1), lost.Load())

	totals := e.totals(t)
	assert.Equal(t, int64(2), totals.Revision)
	assert.Equal(t, 1, totals.HostCount)
	object, _, err := e.s.Read(ctx, e.space, "records/r.sw")
	require.NoError(t, err)
	assert.Equal(t, winner.Load(), object.ETag)
}

// Writes to different files all land, each with a revision of its own, and the changes list misses none of them even
// when read page by page while they happen.
func TestChangesMissNothingUnderConcurrentWrites(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	const files = 60
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := range files {
		wg.Go(func() {
			name := fmt.Sprintf("a%d", i)
			_, err := e.s.Write(ctx, e.space, "records/"+name+".sw", recordBytes(t, name, "other", true, "x"),
				syncservice.Condition{IfAbsent: true})
			assert.NoError(t, err)
		})
	}
	go func() { wg.Wait(); close(done) }()

	seen := map[string]bool{}
	var since int64
	finished := false
	for !finished {
		select {
		case <-done:
			finished = true
		default:
		}
		for {
			changes, err := e.s.ChangesSince(ctx, e.space, since, 7)
			require.NoError(t, err)
			for _, object := range changes.Objects {
				assert.Greater(t, object.Revision, since)
				seen[object.Name] = true
			}
			since = changes.Revision
			if !changes.More {
				break
			}
		}
	}
	assert.Len(t, seen, files)
	assert.Equal(t, int64(files), since)
	assert.Equal(t, files, e.totals(t).RecordCount)
}

func TestChangesReportRemovalsToThoseWhoHadTheFile(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	_, err := e.s.Write(ctx, e.space, "pairing/x.json", []byte("{}"), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "pairing/y.json", []byte("[]"), syncservice.Condition{})
	require.NoError(t, err)
	require.NoError(t, e.s.Remove(ctx, e.space, "pairing/x.json", syncservice.Condition{}))

	changes, err := e.s.ChangesSince(ctx, e.space, 2, 100)
	require.NoError(t, err)
	require.Len(t, changes.Objects, 1)
	assert.Equal(t, "pairing/x.json", changes.Objects[0].Name)
	assert.False(t, changes.Objects[0].IsLive())

	fresh, err := e.s.ChangesSince(ctx, e.space, 0, 100)
	require.NoError(t, err)
	require.Len(t, fresh.Objects, 1, "a client starting from nothing has no copy to drop")
	assert.Equal(t, "pairing/y.json", fresh.Objects[0].Name)
}

func TestSpaceLimit(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	cfg := config.Current()
	old := cfg.Sync.MaxSpaceBytes
	cfg.Sync.MaxSpaceBytes = 100
	t.Cleanup(func() { cfg.Sync.MaxSpaceBytes = old })

	big, err := e.s.Write(ctx, e.space, "attachments/a.sw", make([]byte, 80), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "attachments/b.sw", make([]byte, 21), syncservice.Condition{})
	require.ErrorIs(t, err, hperrors.ErrSyncSpaceFull)
	assert.Equal(t, int64(80), e.totals(t).BytesUsed)

	// Past the limit by a lowered setting: what shrinks is let through, and what grows is not.
	cfg.Sync.MaxSpaceBytes = 50
	_, err = e.s.Write(ctx, e.space, "attachments/a.sw", make([]byte, 81), syncservice.Condition{})
	require.ErrorIs(t, err, hperrors.ErrSyncSpaceFull)
	_, err = e.s.Write(ctx, e.space, "attachments/a.sw", make([]byte, 60), syncservice.Condition{IfMatch: big.ETag})
	require.NoError(t, err)
	assert.Equal(t, int64(60), e.totals(t).BytesUsed)
	require.NoError(t, e.s.Remove(ctx, e.space, "attachments/a.sw", syncservice.Condition{}))
	assert.Equal(t, int64(0), e.totals(t).BytesUsed)
}

func (e *testEnv) blobExists(t *testing.T, name, etag string) bool {
	t.Helper()
	_, err := e.store.Get(context.Background(), entity.SyncBlobKey(e.space, name, etag))
	if err != nil {
		require.ErrorIs(t, err, blobstore.ErrNotFound)
		return false
	}
	return true
}

func (e *testEnv) makeTrashDue(t *testing.T) {
	t.Helper()
	_, err := e.s.db.NewUpdate().Model((*entity.SyncTrash)(nil)).
		Set("delete_after = ?", timeutil.NowUTC().Add(-time.Minute)).
		Where("space_id = ?", e.space).Exec(context.Background())
	require.NoError(t, err)
}

// Replaced content waits out its delay, then goes, unless the same content was written back meanwhile.
func TestCleaningTheTrash(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	one, err := e.s.Write(ctx, e.space, "attachments/a.sw", []byte("one"), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "attachments/a.sw", []byte("two"), syncservice.Condition{})
	require.NoError(t, err)

	require.NoError(t, e.s.emptyTrash(ctx))
	assert.True(t, e.blobExists(t, "attachments/a.sw", one.ETag), "its delay has not passed: a read may be under way")

	e.makeTrashDue(t)
	require.NoError(t, e.s.emptyTrash(ctx))
	assert.False(t, e.blobExists(t, "attachments/a.sw", one.ETag))

	// two is replaced by one's content, then two's content is written back: two is both in the trash and current.
	back, err := e.s.Write(ctx, e.space, "attachments/a.sw", []byte("one"), syncservice.Condition{})
	require.NoError(t, err)
	two, err := e.s.Write(ctx, e.space, "attachments/a.sw", []byte("two"), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "attachments/a.sw", []byte("one"), syncservice.Condition{})
	require.NoError(t, err)
	e.makeTrashDue(t)
	require.NoError(t, e.s.emptyTrash(ctx))
	assert.True(t, e.blobExists(t, "attachments/a.sw", back.ETag), "the line names it")
	assert.False(t, e.blobExists(t, "attachments/a.sw", two.ETag))
	_, data, err := e.s.Read(ctx, e.space, "attachments/a.sw")
	require.NoError(t, err)
	assert.Equal(t, []byte("one"), data)
}

// A write that loses leaves no content behind for good.
func TestLosingWriteIsCleaned(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	cfg := config.Current()
	old := cfg.Sync.MaxSpaceBytes
	cfg.Sync.MaxSpaceBytes = 10
	t.Cleanup(func() { cfg.Sync.MaxSpaceBytes = old })

	data := make([]byte, 11)
	_, err := e.s.Write(ctx, e.space, "attachments/a.sw", data, syncservice.Condition{})
	require.ErrorIs(t, err, hperrors.ErrSyncSpaceFull)
	assert.True(t, e.blobExists(t, "attachments/a.sw", etagOf(data)), "stored before the line decided")
	e.makeTrashDue(t)
	require.NoError(t, e.s.emptyTrash(ctx))
	assert.False(t, e.blobExists(t, "attachments/a.sw", etagOf(data)))
}

// The sweep removes old content nothing knows of, and leaves everything else.
func TestSweep(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	live, err := e.s.Write(ctx, e.space, "records/k.sw", recordBytes(t, "k", "host", true, "x"), syncservice.Condition{})
	require.NoError(t, err)
	stray := entity.SyncBlobKey(e.space, "records/k.sw", "stray")
	require.NoError(t, e.store.Put(ctx, stray, []byte("left by a crash")))
	young := entity.SyncBlobKey(e.space, "records/k.sw", "young")
	require.NoError(t, e.store.Put(ctx, young, []byte("a write under way")))

	local, ok := e.store.(interface{ Root() string })
	require.True(t, ok)
	old := timeutil.NowUTC().Add(-2 * sweepMinAge)
	for _, key := range []string{stray, entity.SyncBlobKey(e.space, "records/k.sw", live.ETag)} {
		require.NoError(t, os.Chtimes(filepath.Join(local.Root(), key), old, old))
	}

	require.NoError(t, e.s.sweep(ctx))
	assert.False(t, e.blobExists(t, "records/k.sw", "stray"))
	assert.True(t, e.blobExists(t, "records/k.sw", "young"), "too young to tell")
	assert.True(t, e.blobExists(t, "records/k.sw", live.ETag), "the line names it")
}

// One pass of the cleaning, under its lock, does both jobs. An instance that finds the lock taken leaves the work.
func TestCleanOnce(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	one, err := e.s.Write(ctx, e.space, "attachments/a.sw", []byte("one"), syncservice.Condition{})
	require.NoError(t, err)
	_, err = e.s.Write(ctx, e.space, "attachments/a.sw", []byte("two"), syncservice.Condition{})
	require.NoError(t, err)
	e.makeTrashDue(t)

	lock, acquired, err := dblock.TryAcquire(ctx, e.s.db, cleanLockName)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NoError(t, e.s.cleanOnce(ctx, false))
	assert.True(t, e.blobExists(t, "attachments/a.sw", one.ETag), "another instance holds the lock")
	require.NoError(t, lock.Release(ctx))

	require.NoError(t, e.s.cleanOnce(ctx, true))
	assert.False(t, e.blobExists(t, "attachments/a.sw", one.ETag))
}

func TestEnsurePersonalSpaceMakesOne(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	space, err := e.s.GetSpace(ctx, e.space)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			again, err := e.s.EnsurePersonalSpace(ctx, space.OwnerUserID)
			if assert.NoError(t, err) {
				assert.Equal(t, space.ID, again.ID)
			}
		})
	}
	wg.Wait()
}

func TestSplitBlobKey(t *testing.T) {
	space, name, etag, ok := splitBlobKey(entity.SyncBlobKey("s1", "records/a.sw", "e1"))
	assert.True(t, ok)
	assert.Equal(t, []string{"s1", "records/a.sw", "e1"}, []string{space, name, etag})
	space, name, etag, ok = splitBlobKey(entity.SyncBlobKey("s1", "format.json", "e1"))
	assert.True(t, ok)
	assert.Equal(t, []string{"s1", "format.json", "e1"}, []string{space, name, etag})
	for _, key := range []string{"spaces/s1/e1", "other/s1/a/e1"} {
		space, name, etag, ok = splitBlobKey(key)
		assert.False(t, ok, key)
		assert.Empty(t, space+name+etag, key)
	}
}
