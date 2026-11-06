package identityuc

import (
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
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

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/identity"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/identityuc/identitydto"
)

type fakeAudit struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (f *fakeAudit) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

// These need a real PostgreSQL with the migrations applied: point SHW_TEST_PG_DSN at one. Each test has a schema of its
// own.
func newUC(t *testing.T) (*UC, *database.DB, *fakeAudit) {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping identity test")
	}
	ctx := context.Background()
	schema := "identity_" + strings.ToLower(gofn.Must(ulid.NewStringULID()))
	open := func(conf *pgx.ConnConfig) *bun.DB { return bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New()) }
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	admin := open(conf)
	t.Cleanup(func() { _ = admin.Close() })
	for _, stmt := range []string{
		"CREATE SCHEMA " + schema,
		"CREATE TABLE " + schema + ".users (LIKE public.users INCLUDING ALL)",
		"CREATE TABLE " + schema + ".user_identities (LIKE public.user_identities INCLUDING ALL)",
		"CREATE TABLE " + schema + ".collection_key_wraps (LIKE public.collection_key_wraps INCLUDING ALL)",
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
	audit := &fakeAudit{}
	return New(db, repository.NewUserIdentityRepo(), repository.NewCollectionRepo(), audit), db, audit
}

func person(t *testing.T, db *database.DB) *basedto.Auth {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	_, err := db.NewInsert().Model(&entity.User{
		ID: id, Username: id, Role: base.UserRoleMember, Status: base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
	}).Exec(context.Background())
	require.NoError(t, err)
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: id, Username: id}}}
}

// newIdentity is what an app publishes: fresh keys, the X25519 one signed.
func newIdentity(t *testing.T) *identitydto.PutIdentityReq {
	t.Helper()
	x, err := ecdh.X25519().GenerateKey(rand.Reader)
	require.NoError(t, err)
	edPub, edSecret, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	xPub := x.PublicKey().Bytes()
	msg := append(append([]byte("shellwing identity signature v1"), xPub...), edPub...)
	return &identitydto.PutIdentityReq{X25519Public: xPub, Ed25519Public: edPub, Signature: ed25519.Sign(edSecret, msg)}
}

func TestAPersonPublishesTheirIdentity(t *testing.T) {
	uc, db, audit := newUC(t)
	ctx, me := context.Background(), person(t, db)

	got, err := uc.GetMine(ctx, me)
	require.NoError(t, err)
	assert.Nil(t, got.Data, "none yet")

	first := newIdentity(t)
	put, err := uc.PutMine(ctx, me, first)
	require.NoError(t, err)
	assert.Equal(t, identity.Fingerprint(first.X25519Public, first.Ed25519Public), put.Data.Fingerprint)
	got, err = uc.GetMine(ctx, me)
	require.NoError(t, err)
	assert.Equal(t, first.X25519Public, got.Data.X25519Public)
	require.Len(t, audit.entries, 1)

	// A collection's key wrapped for the first.
	_, err = db.NewInsert().Model(&entity.CollectionKeyWrap{
		CollectionID: "c1", UserID: me.UserID(), KeyVersion: 1, RecipientX25519: first.X25519Public,
		EphemeralPublic: []byte{1}, SealedKey: []byte{1}, Commitment: []byte{1}, WrappedBy: me.UserID(),
		WrapperX25519: []byte{1}, WrapperEd25519: []byte{1}, Signature: []byte{1},
	}).Exec(ctx)
	require.NoError(t, err)
	wraps := func() int {
		n, err := db.NewSelect().Model((*entity.CollectionKeyWrap)(nil)).Where("user_id = ?", me.UserID()).Count(ctx)
		require.NoError(t, err)
		return int(n)
	}

	// The same again changes nothing; another takes its place, with the one before in the log, and the keys wrapped
	// for the one before go.
	_, err = uc.PutMine(ctx, me, first)
	require.NoError(t, err)
	require.Len(t, audit.entries, 1)
	assert.Equal(t, 1, wraps())
	second := newIdentity(t)
	_, err = uc.PutMine(ctx, me, second)
	require.NoError(t, err)
	require.Len(t, audit.entries, 2)
	assert.Contains(t, audit.entries[1].Detail, put.Data.Fingerprint)
	assert.Contains(t, audit.entries[1].Detail, `"wrapsDropped":1`)
	assert.Equal(t, 0, wraps())
	got, err = uc.GetMine(ctx, me)
	require.NoError(t, err)
	assert.Equal(t, second.Ed25519Public, got.Data.Ed25519Public)

	// Each person has their own.
	other := person(t, db)
	got, err = uc.GetMine(ctx, other)
	require.NoError(t, err)
	assert.Nil(t, got.Data)
}

func TestAnIdentityNotSignedIsRefused(t *testing.T) {
	uc, db, _ := newUC(t)
	ctx, me := context.Background(), person(t, db)
	forged := newIdentity(t)
	forged.X25519Public = newIdentity(t).X25519Public
	_, err := uc.PutMine(ctx, me, forged)
	require.ErrorIs(t, err, hperrors.ErrIdentityInvalid)
	short := newIdentity(t)
	short.Ed25519Public = short.Ed25519Public[:31]
	_, err = uc.PutMine(ctx, me, short)
	require.ErrorIs(t, err, hperrors.ErrIdentityInvalid)
	got, err := uc.GetMine(ctx, me)
	require.NoError(t, err)
	assert.Nil(t, got.Data)
}
