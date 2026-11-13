package datakeyserviceimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
	"github.com/hivepaas/shellwing-server/pkg/kms"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/datakeyservice"
)

const testAppSecret = "the-app-secret-under-test"

type env struct {
	svc datakeyservice.Service
	db  *database.DB
}

// These need a real PostgreSQL with the migrations applied. Point SHW_TEST_PG_DSN at one to enable them. Each test has
// a schema of its own.
func newEnv(t *testing.T) *env {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping data key test")
	}

	ctx := context.Background()
	schema := "dek_" + strings.ToLower(gofn.Must(ulid.NewStringULID()))
	open := func(conf *pgx.ConnConfig) *bun.DB { return bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New()) }
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	admin := open(conf)
	t.Cleanup(func() { _ = admin.Close() })
	for _, stmt := range []string{
		"CREATE SCHEMA " + schema,
		"CREATE TABLE " + schema + ".encryption_keys (LIKE public.encryption_keys INCLUDING ALL)",
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

	// The active key is process state, as it is in the app. Put it back as it was
	// so a test that installs one does not decide what the next test reads.
	previous := datakey.Active()
	t.Cleanup(func() { datakey.SetActive(previous) })

	return &env{
		svc: New(repository.NewEncryptionKeyRepo(), logging.GlobalLogger()),
		db:  &database.DB{DB: bunDB},
	}
}

// useConfig installs the configuration Load reads, and puts back what was there.
func useConfig(t *testing.T, kmsCfg config.KMS) {
	t.Helper()
	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })
	config.SetCurrent(&config.Config{Secret: testAppSecret, KMS: kmsCfg})
}

// stored is the wrapped key as the database holds it.
func (e *env) stored(t *testing.T) string {
	t.Helper()
	row, err := repository.NewEncryptionKeyRepo().GetActive(context.Background(), e.db)
	require.NoError(t, err)
	require.NotNil(t, row)
	return row.WrappedKey
}

// fakeTransit answers as Vault's transit engine does. Its cipher is the identity:
// what is under test is which provider wraps the key and what is stored, not
// anybody's AES.
func fakeTransit(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var body map[string]string
		require.NoError(t, json.Unmarshal(raw, &body))

		answer := map[string]any{}
		switch {
		case strings.Contains(r.URL.Path, "/encrypt/"):
			answer["data"] = map[string]string{"ciphertext": "vault:v1:" + body["plaintext"]}
		case strings.Contains(r.URL.Path, "/decrypt/"):
			answer["data"] = map[string]string{
				"plaintext": strings.TrimPrefix(body["ciphertext"], "vault:v1:"),
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(answer))
	}))
	t.Cleanup(server.Close)
	return server
}

func transitConfig(server *httptest.Server) config.KMS {
	return config.KMS{
		Provider:     kms.ProviderVaultTransit,
		VaultAddress: server.URL,
		VaultToken:   "a-server-token",
		VaultMount:   "transit",
		VaultKey:     "shellwing-dek",
	}
}

func TestTheKeyIsMadeOnTheFirstStartAndReadBackOnTheNext(t *testing.T) {
	e := newEnv(t)
	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	ctx := context.Background()

	require.NoError(t, e.svc.Load(ctx, e.db))
	first := datakey.Active()
	require.NotNil(t, first)
	sealed, err := first.Seal("a stored secret")
	require.NoError(t, err)
	assert.Equal(t, kms.ProviderLocal, kms.ProviderOf(e.stored(t)),
		"wrapped with the app secret, in the form every existing install has")

	// A second start, as a restart is.
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))
	opened, err := datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "a stored secret", opened, "the same key, so what it sealed is still readable")
}

func TestAnAppSecretThatDoesNotOpenTheKeyStopsTheStart(t *testing.T) {
	e := newEnv(t)
	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	ctx := context.Background()
	require.NoError(t, e.svc.Load(ctx, e.db))

	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })
	config.SetCurrent(&config.Config{Secret: "somebody-changed-it", KMS: config.KMS{Provider: kms.ProviderLocal}})

	// As a fresh process is: nothing installed yet.
	datakey.SetActive(nil)
	err := e.svc.Load(ctx, e.db)
	require.ErrorIs(t, err, hperrors.ErrPreconditionFailed, "unreadable settings later is the worse outcome")
	assert.Nil(t, datakey.Active(), "and nothing is installed")
}

// Changing provider is a change to the configuration, applied on the next start.
// It must move the key without changing it: what was sealed before stays
// readable, because the values themselves are never re-encrypted.
func TestTheKeyMovesToTheServiceTheConfigurationNames(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	require.NoError(t, e.svc.Load(ctx, e.db))
	sealed, err := datakey.Active().Seal("sealed before the move")
	require.NoError(t, err)
	wasStored := e.stored(t)

	server := fakeTransit(t)
	useConfig(t, transitConfig(server))
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))

	nowStored := e.stored(t)
	assert.Equal(t, kms.ProviderVaultTransit, kms.ProviderOf(nowStored), "the row says who wraps it now")
	assert.NotEqual(t, wasStored, nowStored)
	opened, err := datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "sealed before the move", opened, "the key did not change, only its wrapping")

	// And the next start reads it from the service.
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))
	opened, err = datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "sealed before the move", opened)
	assert.Equal(t, nowStored, e.stored(t), "and leaves the row alone, having nothing to move")
}

// Back again, because an operator who turns a service on must be able to turn it
// off: the key is theirs either way.
func TestTheKeyMovesBackToTheAppSecret(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	server := fakeTransit(t)

	useConfig(t, transitConfig(server))
	require.NoError(t, e.svc.Load(ctx, e.db))
	sealed, err := datakey.Active().Seal("sealed under the service")
	require.NoError(t, err)
	assert.Equal(t, kms.ProviderVaultTransit, kms.ProviderOf(e.stored(t)))

	// The engine's settings stay: it is the only thing that can open the key, and
	// moving away from it is a call to it. Only the provider changes.
	movingBack := transitConfig(server)
	movingBack.Provider = kms.ProviderLocal
	useConfig(t, movingBack)
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))
	assert.Equal(t, kms.ProviderLocal, kms.ProviderOf(e.stored(t)))
	opened, err := datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "sealed under the service", opened)

	// Now that the key is the app secret's again, the engine's settings can go.
	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))
	opened, err = datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "sealed under the service", opened)
}

// The trap an operator will meet: turning a service off is not only a change of
// provider, because the key still has to be opened by the service that wrapped
// it. Dropping its settings in the same change leaves a server that cannot start,
// so it says exactly that instead of a derivation that merely fails.
func TestMovingAwayFromAServiceStillNeedsItsSettings(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	server := fakeTransit(t)

	useConfig(t, transitConfig(server))
	require.NoError(t, e.svc.Load(ctx, e.db))
	wasStored := e.stored(t)

	// The provider changed and the engine's settings went with it.
	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	datakey.SetActive(nil)
	err := e.svc.Load(ctx, e.db)
	require.ErrorIs(t, err, hperrors.ErrPreconditionFailed)
	assert.Equal(t, wasStored, e.stored(t), "and the key is left where the engine can still open it")
}

// A key wrapped by a service this server can no longer reach is the restored
// backup case, and the one where carrying on would be worst: the stored values
// are unreadable, and a server that started would write new ones beside them
// under a different key.
func TestAServiceThatCannotBeReachedStopsTheStart(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	server := fakeTransit(t)

	useConfig(t, transitConfig(server))
	require.NoError(t, e.svc.Load(ctx, e.db))
	wasStored := e.stored(t)

	server.Close() // as a sealed Vault, or one behind a network that is gone
	datakey.SetActive(nil)
	err := e.svc.Load(ctx, e.db)
	require.Error(t, err)
	assert.Nil(t, datakey.Active(), "and nothing is installed")
	assert.Equal(t, wasStored, e.stored(t), "and the row is left as it was, for a start that can reach it")
}

func TestChangingTheAppSecretRewrapsTheKeyAndNothingElse(t *testing.T) {
	e := newEnv(t)
	useConfig(t, config.KMS{Provider: kms.ProviderLocal})
	ctx := context.Background()
	require.NoError(t, e.svc.Load(ctx, e.db))
	sealed, err := datakey.Active().Seal("a stored secret")
	require.NoError(t, err)
	wasStored := e.stored(t)

	require.Error(t, e.svc.Rewrap(ctx, e.db, "not-the-current-one", "a-new-secret"),
		"the current secret is what proves the caller may change it")
	assert.Equal(t, wasStored, e.stored(t))

	require.NoError(t, e.svc.Rewrap(ctx, e.db, testAppSecret, "a-new-secret"))
	nowStored := e.stored(t)
	assert.NotEqual(t, wasStored, nowStored)
	assert.Equal(t, kms.ProviderLocal, kms.ProviderOf(nowStored))

	// The key did not change, so what it sealed is still readable - and the new
	// secret is what opens it now.
	opened, err := datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "a stored secret", opened)

	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })
	config.SetCurrent(&config.Config{Secret: "a-new-secret", KMS: config.KMS{Provider: kms.ProviderLocal}})
	datakey.SetActive(nil)
	require.NoError(t, e.svc.Load(ctx, e.db))
	opened, err = datakey.Active().Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "a stored secret", opened)
}

// Rotating the app secret when a service holds the key encryption key protects
// nothing, so it is refused rather than reported as done.
func TestChangingTheAppSecretIsRefusedWhenAServiceHoldsTheKey(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	server := fakeTransit(t)
	useConfig(t, transitConfig(server))
	require.NoError(t, e.svc.Load(ctx, e.db))
	wasStored := e.stored(t)

	err := e.svc.Rewrap(ctx, e.db, testAppSecret, "a-new-secret")
	require.Error(t, err)
	assert.Equal(t, wasStored, e.stored(t), "and leaves the key wrapped as it was")
}

// The key reaches the service base64, which is what transit takes, and the
// stored value is the ciphertext it answered with under this package's marker.
func TestWhatIsStoredForAServiceWrappedKey(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	server := fakeTransit(t)
	useConfig(t, transitConfig(server))
	require.NoError(t, e.svc.Load(ctx, e.db))

	stored := e.stored(t)
	require.True(t, strings.HasPrefix(stored, "hpkms:vault-transit:vault:v1:"), stored)
	encoded := strings.TrimPrefix(stored, "hpkms:vault-transit:vault:v1:")
	raw, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	assert.Len(t, raw, datakey.KeyLen, "a whole data encryption key and nothing else")
}
