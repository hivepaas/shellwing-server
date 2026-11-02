package licenseserviceimpl

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"
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
	"github.com/hivepaas/shellwing-server/pkg/licensesig"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/licenseservice"
)

// The installation the keys in pkg/licensesig/testdata name.
const testInstallation = "q9xk2m4j7n1p0r5s8t3v6w2yz"

func testKey(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../../../pkg/licensesig/testdata/" + name + ".txt")
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}

func devConfig(limit int) {
	cfg := &config.Config{}
	cfg.DevMode.Enabled = true
	cfg.DevMode.UserLimit = limit
	config.SetCurrent(cfg)
}

func TestNewInstallationIDIsOneThePortalTakes(t *testing.T) {
	re := regexp.MustCompile(`^[a-z0-9]{25}$`)
	assert.Regexp(t, re, newInstallationID())
	assert.NotEqual(t, newInstallationID(), newInstallationID())
}

func TestEvaluate(t *testing.T) {
	config.SetCurrent(&config.Config{})
	expires := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	key := func() *licensesig.Payload {
		return &licensesig.Payload{Plan: "team-10", Limits: map[string]int{licensesig.LimitUsers: 10},
			Installations: []string{"here"}, ExpiresAt: expires, GraceDays: 30}
	}
	free := func(state *licenseservice.State, status base.LicenseStatus) {
		t.Helper()
		assert.Equal(t, status, state.Status)
		assert.Equal(t, base.FreeUserLimit, state.MaxUsers, status)
		assert.Empty(t, state.Plan, status)
	}
	before := expires.Add(-time.Hour)

	free(evaluate("here", nil, nil, false, before), base.LicenseStatusNone)
	free(evaluate("here", nil, errors.New("no"), true, before), base.LicenseStatusInvalid)
	free(evaluate("elsewhere", key(), nil, true, before), base.LicenseStatusOtherInstallation)
	free(evaluate("here", key(), nil, true, expires.Add(31*24*time.Hour)), base.LicenseStatusExpired)

	state := evaluate("here", key(), nil, true, before)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
	assert.Equal(t, 10, state.MaxUsers)
	assert.Equal(t, "team-10", state.Plan)

	state = evaluate("here", key(), nil, true, expires.Add(29*24*time.Hour))
	assert.Equal(t, base.LicenseStatusGrace, state.Status, "still in force through its grace")
	assert.Equal(t, 10, state.MaxUsers)

	trial := key()
	trial.Trial = true
	assert.Equal(t, base.LicenseStatusTrial, evaluate("here", trial, nil, true, before).Status)

	// A limit left out of a key is no limit.
	open := key()
	open.Limits = nil
	assert.Equal(t, base.NoUserLimit, evaluate("here", open, nil, true, before).MaxUsers)
}

func TestTheDevelopmentLimitWorksOnlyInDevelopment(t *testing.T) {
	cfg := &config.Config{}
	cfg.DevMode.UserLimit = 100
	config.SetCurrent(cfg)
	assert.Equal(t, base.FreeUserLimit, freeUserLimit(), "set by mistake on a real server")
	devConfig(100)
	assert.Equal(t, 100, freeUserLimit())
	devConfig(0)
	assert.Equal(t, base.FreeUserLimit, freeUserLimit())
}

// These need a real PostgreSQL with the migrations applied: what they check is that the lock holds under real
// concurrency. Point SHW_TEST_PG_DSN at one to enable them. Each test has a schema of its own, with a users table
// nothing else writes to: the limit counts them all.
func newEnv(t *testing.T) (*service, *database.DB) {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping license test")
	}
	ctx := context.Background()
	schema := "lic_" + strings.ToLower(gofn.Must(ulid.NewStringULID()))
	open := func(conf *pgx.ConnConfig) *bun.DB {
		return bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New())
	}
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	admin := open(conf)
	t.Cleanup(func() { _ = admin.Close() })
	for _, stmt := range []string{
		"CREATE SCHEMA " + schema,
		"CREATE TABLE " + schema + ".users (LIKE public.users INCLUDING ALL)",
		"CREATE TABLE " + schema + ".license_state (LIKE public.license_state INCLUDING ALL)",
	} {
		_, err = admin.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}
	t.Cleanup(func() { _, _ = admin.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE") })

	scoped, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	scoped.RuntimeParams["search_path"] = schema
	bunDB := open(scoped)
	t.Cleanup(func() { _ = bunDB.Close() })
	db := &database.DB{DB: bunDB}

	s, ok := New(repository.NewLicenseRepo()).(*service)
	require.True(t, ok)
	require.NoError(t, s.EnsureInstallation(ctx, db))
	return s, db
}

func addUser(t *testing.T, db database.IDB, status base.UserStatus) *entity.User {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	user := &entity.User{ID: id, Username: id, Role: base.UserRoleMember, Status: status,
		SecurityOption: base.UserSecurityPasswordOnly}
	_, err := db.NewInsert().Model(user).Exec(context.Background())
	require.NoError(t, err)
	return user
}

func check(ctx context.Context, s *service, db *database.DB) error {
	return transaction.Execute(ctx, db, func(tx database.Tx) error {
		return s.CheckUserSeat(ctx, tx)
	})
}

func TestTheFreeLimitAndWhatHoldsASeat(t *testing.T) {
	devConfig(0)
	s, db := newEnv(t)
	ctx := context.Background()

	require.NoError(t, check(ctx, s, db))
	addUser(t, db, base.UserStatusActive)
	require.NoError(t, check(ctx, s, db))
	pending := addUser(t, db, base.UserStatusPending)

	err := check(ctx, s, db)
	require.ErrorIs(t, err, hperrors.ErrPlanLimitReached, "an invited user holds a seat")

	// A disabled user holds none.
	pending.Status = base.UserStatusDisabled
	_, err = db.NewUpdate().Model(pending).Column("status").WherePK().Exec(ctx)
	require.NoError(t, err)
	require.NoError(t, check(ctx, s, db))

	// Nor a deleted one.
	pending.Status = base.UserStatusActive
	_, err = db.NewUpdate().Model(pending).Column("status").WherePK().Exec(ctx)
	require.NoError(t, err)
	require.ErrorIs(t, check(ctx, s, db), hperrors.ErrPlanLimitReached)
	_, err = db.NewDelete().Model(pending).WherePK().Exec(ctx)
	require.NoError(t, err)
	require.NoError(t, check(ctx, s, db))
}

// Of many creations at once, as many get in as there are seats, and no more.
func TestTwoAtOnceCannotTakeTheLastSeat(t *testing.T) {
	devConfig(0)
	s, db := newEnv(t)
	ctx := context.Background()

	var made, refused atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			err := transaction.Execute(ctx, db, func(tx database.Tx) error {
				if err := s.CheckUserSeat(ctx, tx); err != nil {
					return err //nolint:wrapcheck // compared below
				}
				id := gofn.Must(ulid.NewStringULID())
				_, err := tx.NewInsert().Model(&entity.User{ID: id, Username: id, Role: base.UserRoleMember,
					Status: base.UserStatusPending, SecurityOption: base.UserSecurityPasswordOnly}).Exec(ctx)
				return err //nolint:wrapcheck // compared below
			})
			switch {
			case err == nil:
				made.Add(1)
			case errors.Is(err, hperrors.ErrPlanLimitReached):
				refused.Add(1)
			default:
				assert.NoError(t, err)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, int32(base.FreeUserLimit), made.Load())
	assert.Equal(t, int32(12-base.FreeUserLimit), refused.Load())
	count, err := s.CountUserSeats(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, base.FreeUserLimit, count)
}

func setInstallation(t *testing.T, db *database.DB, id string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "UPDATE license_state SET installation_id = ?", id)
	require.NoError(t, err)
}

func TestAKeyLiftsTheLimitAndRemovingItBringsItBack(t *testing.T) {
	devConfig(0)
	s, db := newEnv(t)
	ctx := context.Background()
	setInstallation(t, db, testInstallation)
	for range 3 {
		addUser(t, db, base.UserStatusActive)
	}
	require.ErrorIs(t, check(ctx, s, db), hperrors.ErrPlanLimitReached)

	state, err := s.Install(ctx, db, "  "+testKey(t, "team-10")+"\n")
	require.NoError(t, err)
	assert.Equal(t, 10, state.MaxUsers)
	assert.Equal(t, "team-10", state.Plan)
	assert.Equal(t, "Acme", state.Key.Customer.Name)
	require.NoError(t, check(ctx, s, db))

	state, err = s.Current(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, 10, state.MaxUsers, "kept, and read back")

	// Removing it stops creating, and nothing else: all three are still there.
	state, err = s.Remove(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, base.LicenseStatusNone, state.Status)
	assert.Equal(t, base.FreeUserLimit, state.MaxUsers)
	require.ErrorIs(t, check(ctx, s, db), hperrors.ErrPlanLimitReached)
	count, err := s.CountUserSeats(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestInstallRefuses(t *testing.T) {
	devConfig(0)
	s, db := newEnv(t)
	ctx := context.Background()

	_, err := s.Install(ctx, db, "not a key")
	require.ErrorIs(t, err, hperrors.ErrLicenseKeyInvalid)
	_, err = s.Install(ctx, db, testKey(t, "hivepaas-key"))
	require.ErrorIs(t, err, hperrors.ErrLicenseKeyInvalid, "another product's")
	_, err = s.Install(ctx, db, testKey(t, "team-10"))
	require.ErrorIs(t, err, hperrors.ErrLicenseKeyOtherInstallation, "this server has another ID")

	state, err := s.Current(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, base.LicenseStatusNone, state.Status, "nothing was kept")

	// Outside development, a key the public development pair signed is no key.
	setInstallation(t, db, testInstallation)
	config.SetCurrent(&config.Config{})
	_, err = s.Install(ctx, db, testKey(t, "team-10"))
	require.ErrorIs(t, err, hperrors.ErrLicenseKeyInvalid)
}

// A server whose key stops verifying, as when moved between environments, falls back to the free limit, and starts.
func TestAKeyThatNoLongerVerifiesPutsNothingInForce(t *testing.T) {
	devConfig(0)
	s, db := newEnv(t)
	ctx := context.Background()
	setInstallation(t, db, testInstallation)
	_, err := s.Install(ctx, db, testKey(t, "team-10"))
	require.NoError(t, err)

	config.SetCurrent(&config.Config{})
	state, err := s.Current(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, base.LicenseStatusInvalid, state.Status)
	assert.Equal(t, base.FreeUserLimit, state.MaxUsers)
}
