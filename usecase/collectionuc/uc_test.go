package collectionuc

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
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/collectionkey"
	"github.com/hivepaas/shellwing-server/pkg/identity"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc/collectiondto"
	"github.com/hivepaas/shellwing-server/usecase/syncuc"
)

type fakeAudit struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (f *fakeAudit) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeAudit) types() []base.AuditLogType {
	out := make([]base.AuditLogType, 0, len(f.entries))
	for _, e := range f.entries {
		out = append(out, e.Type)
	}
	return out
}

// fakeSync is the sync service as far as access goes: the spaces are the database's, and hold nothing.
type fakeSync struct {
	syncservice.Service
	db     *database.DB
	spaces repository.SpaceRepo
}

func (f *fakeSync) GetSpace(ctx context.Context, spaceID string) (*entity.Space, error) {
	return f.spaces.GetByID(ctx, f.db, spaceID) //nolint:wrapcheck // a test's
}

func (f *fakeSync) EnsurePersonalSpace(_ context.Context, userID string) (*entity.Space, error) {
	return &entity.Space{ID: "personal-" + userID, Type: base.SpaceTypePersonal, OwnerUserID: userID}, nil
}

func (f *fakeSync) GetTotals(_ context.Context, spaceID string) (*entity.SyncSpace, error) {
	return &entity.SyncSpace{SpaceID: spaceID}, nil
}

type env struct {
	uc    *UC
	sync  *syncuc.UC
	db    *database.DB
	audit *fakeAudit
	org   string
	names map[string]string
}

// These need a real PostgreSQL with the migrations applied. Point SHW_TEST_PG_DSN at one to enable them. Each test has
// a schema of its own.
func newEnv(t *testing.T) *env {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping collection test")
	}
	if config.Current() == nil {
		config.SetCurrent(&config.Config{})
		t.Cleanup(func() { config.SetCurrent(nil) })
	}

	ctx := context.Background()
	schema := "collection_" + strings.ToLower(gofn.Must(ulid.NewStringULID()))
	open := func(conf *pgx.ConnConfig) *bun.DB { return bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New()) }
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	admin := open(conf)
	t.Cleanup(func() { _ = admin.Close() })
	stmts := []string{"CREATE SCHEMA " + schema}
	for _, table := range []string{"organizations", "org_members", "users", "org_groups", "org_group_members",
		"spaces", "sync_spaces", "sync_objects", "sync_trash", "collections", "collection_grants",
		"collection_key_wraps", "user_identities"} {
		stmts = append(stmts, "CREATE TABLE "+schema+"."+table+" (LIKE public."+table+" INCLUDING ALL)")
	}
	for _, stmt := range stmts {
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
	orgRepo, collectionRepo, spaceRepo := repository.NewOrganizationRepo(), repository.NewCollectionRepo(),
		repository.NewSpaceRepo()
	e := &env{
		uc: New(db, orgRepo, repository.NewOrgGroupRepo(), collectionRepo, spaceRepo, repository.NewSyncRepo(),
			repository.NewUserIdentityRepo(), audit),
		sync:  syncuc.New(db, &fakeSync{db: db, spaces: spaceRepo}, orgRepo, collectionRepo),
		db:    db,
		audit: audit,
		names: map[string]string{},
	}
	e.org = gofn.Must(ulid.NewStringULID())
	_, err = db.NewInsert().Model(&entity.Organization{
		ID: e.org, Slug: strings.ToLower(e.org), Name: "Org", Status: base.OrgStatusActive, UpdateVer: 1,
	}).Exec(ctx)
	require.NoError(t, err)
	return e
}

// person is a member of the organization, with the secrets of an identity once they publish one.
type person struct {
	auth   *basedto.Auth
	x25519 *ecdh.PrivateKey
	ed     ed25519.PrivateKey
}

func (p *person) id() string {
	return p.auth.UserID()
}

func (e *env) add(t *testing.T, name string, role base.OrgRole) *person {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	ctx := context.Background()
	_, err := e.db.NewInsert().Model(&entity.User{
		ID: id, Username: name, Role: base.UserRoleMember, Status: base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = e.db.NewInsert().Model(&entity.OrgMember{
		OrgID: e.org, UserID: id, Role: role, Status: base.OrgMemberStatusActive,
	}).Exec(ctx)
	require.NoError(t, err)
	e.names[id] = name
	return &person{auth: &basedto.Auth{User: &basedto.User{
		User:       &entity.User{ID: id, Username: name},
		AuthClaims: &jwtsession.AuthClaims{UID: "device-" + name, UserID: id, Scope: jwtsession.ScopeSync},
	}}}
}

// publish gives the person an identity, as their app does.
func (e *env) publish(t *testing.T, p *person) {
	t.Helper()
	var err error
	p.x25519, err = ecdh.X25519().GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, p.ed, err = ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	x, ed := p.x25519.PublicKey().Bytes(), []byte(p.ed.Public().(ed25519.PublicKey))
	_, err = e.db.NewInsert().Model(&entity.UserIdentity{
		UserID: p.id(), X25519Public: x, Ed25519Public: ed, Signature: make([]byte, 64),
		Fingerprint: identity.Fingerprint(x, ed),
	}).Exec(context.Background())
	require.NoError(t, err)
}

func random(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

// wrap is a key wrapped by one person's app for another: the server checks the signature and for whom, never what is
// sealed, so that is random.
func wrap(t *testing.T, by *person, collectionID string, version int, to *person) *collectiondto.WrapReq {
	t.Helper()
	w := &collectionkey.Wrap{
		CollectionID:    collectionID,
		Version:         version,
		RecipientX25519: to.x25519.PublicKey().Bytes(),
		EphemeralPublic: random(t, collectionkey.KeyLen),
		Sealed:          random(t, collectionkey.SealedLen),
		Commitment:      random(t, collectionkey.KeyLen),
	}
	return &collectiondto.WrapReq{
		UserID:          to.id(),
		RecipientX25519: w.RecipientX25519,
		EphemeralPublic: w.EphemeralPublic,
		SealedKey:       w.Sealed,
		Commitment:      w.Commitment,
		Signature:       ed25519.Sign(by.ed, w.Signed()),
	}
}

func wraps(version int, w ...*collectiondto.WrapReq) *collectiondto.WrapsReq {
	return &collectiondto.WrapsReq{Version: version, Wraps: w}
}

// create makes the collection "Production".
func (e *env) create(t *testing.T, by *person) string {
	t.Helper()
	resp, err := e.uc.Create(context.Background(), by.auth, e.org, &collectiondto.CollectionReq{Name: "Production"})
	require.NoError(t, err)
	return resp.Data.ID
}

func (e *env) grant(t *testing.T, by *person, collectionID string, grants ...*collectiondto.GrantReq) {
	t.Helper()
	_, err := e.uc.SetGrants(context.Background(), by.auth, e.org, collectionID,
		&collectiondto.SetGrantsReq{Grants: grants})
	require.NoError(t, err)
}

func userGrant(p *person, permission base.CollectionPermission) *collectiondto.GrantReq {
	return &collectiondto.GrantReq{UserID: p.id(), Permission: permission}
}

// listed is the collection as the person's list has it, or nil.
func (e *env) listed(t *testing.T, p *person, collectionID string) *collectiondto.CollectionResp {
	t.Helper()
	resp, err := e.uc.List(context.Background(), p.auth, e.org)
	require.NoError(t, err)
	for _, c := range resp.Data {
		if c.ID == collectionID {
			return c
		}
	}
	return nil
}

func TestOwnersAndAdminsMakeCollectionsAndAnAppGivesTheKey(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.add(t, "olivia", base.OrgRoleOwner)
	admin := e.add(t, "adam", base.OrgRoleAdmin)
	member := e.add(t, "mia", base.OrgRoleMember)

	_, err := e.uc.Create(ctx, member.auth, e.org, &collectiondto.CollectionReq{Name: "Production"})
	require.ErrorIs(t, err, hperrors.ErrForbidden, "members do not make collections")
	id := e.create(t, admin)
	_, err = e.uc.Create(ctx, owner.auth, e.org, &collectiondto.CollectionReq{Name: "production"})
	require.ErrorIs(t, err, hperrors.ErrAlreadyExist, "whatever its case")

	// The admin who made it has Manage; the owner has it without a grant; the member does not see it.
	got := e.listed(t, admin, id)
	require.NotNil(t, got)
	assert.Equal(t, base.CollectionPermissionManage, got.Permission)
	assert.Zero(t, got.KeyVersion, "no key yet")
	assert.Zero(t, got.Waiting, "nothing to wait for before the first key")
	assert.Equal(t, base.CollectionPermissionManage, e.listed(t, owner, id).Permission)
	assert.Nil(t, e.listed(t, member, id))
	_, err = e.uc.Get(ctx, member.auth, e.org, id)
	require.ErrorIs(t, err, hperrors.ErrNotFound)

	// The admin's app makes the first key: it needs an identity, and wraps it for every owner who has one.
	_, err = e.uc.CreateKey(ctx, admin.auth, e.org, id, wraps(1))
	require.ErrorIs(t, err, hperrors.ErrIdentityMissing)
	e.publish(t, admin)
	e.publish(t, owner)
	_, err = e.uc.CreateKey(ctx, admin.auth, e.org, id, wraps(1, wrap(t, admin, id, 1, admin)))
	require.ErrorIs(t, err, hperrors.ErrCollectionWrapsMissing, "the owner's")
	_, err = e.uc.CreateKey(ctx, admin.auth, e.org, id, wraps(2, wrap(t, admin, id, 2, admin),
		wrap(t, admin, id, 2, owner)))
	require.ErrorIs(t, err, hperrors.ErrCollectionKeyVersion, "1 first")
	_, err = e.uc.CreateKey(ctx, admin.auth, e.org, id, wraps(1, wrap(t, admin, id, 1, admin),
		wrap(t, admin, id, 1, owner)))
	require.NoError(t, err)
	_, err = e.uc.CreateKey(ctx, owner.auth, e.org, id, wraps(1, wrap(t, owner, id, 1, owner)))
	require.ErrorIs(t, err, hperrors.ErrCollectionKeyVersion, "another app made it first")

	got = e.listed(t, owner, id)
	assert.Equal(t, 1, got.KeyVersion)
	assert.True(t, got.HasKey)
	assert.Zero(t, got.Waiting)

	// The owner's app fetches its wrap, with who made it; the fetch is in the log.
	keys, err := e.uc.GetKeys(ctx, owner.auth, e.org, id)
	require.NoError(t, err)
	require.Len(t, keys.Data.Wraps, 1)
	w := keys.Data.Wraps[0]
	assert.Equal(t, 1, w.Version)
	assert.Equal(t, owner.x25519.PublicKey().Bytes(), w.RecipientX25519)
	assert.Equal(t, "adam", w.WrapperUsername)
	assert.Equal(t, identity.Fingerprint(admin.x25519.PublicKey().Bytes(), admin.ed.Public().(ed25519.PublicKey)),
		w.WrapperFingerprint)
	assert.Contains(t, e.audit.types(), base.AuditLogTypeCollectionKeyFetch)

	// Renamed by whoever manages it; deleted by an admin, and gone for everybody.
	_, err = e.uc.Update(ctx, admin.auth, e.org, id, &collectiondto.CollectionReq{Name: "Prod"})
	require.NoError(t, err)
	assert.Equal(t, "Prod", e.listed(t, owner, id).Name)
	_, err = e.uc.Delete(ctx, admin.auth, e.org, id)
	require.NoError(t, err)
	assert.Nil(t, e.listed(t, owner, id))
	require.ErrorIs(t, e.sync.CheckAccess(ctx, owner.auth, id), hperrors.ErrNotFound)
}

func TestGrantedPeopleWaitUntilSomebodyHoldingTheKeyGivesItThem(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.add(t, "olivia", base.OrgRoleOwner)
	editor := e.add(t, "ed", base.OrgRoleMember)
	viewer := e.add(t, "vic", base.OrgRoleMember)
	stranger := e.add(t, "sam", base.OrgRoleMember)
	for _, p := range []*person{owner, editor, viewer, stranger} {
		e.publish(t, p)
	}
	id := e.create(t, owner)
	_, err := e.uc.CreateKey(ctx, owner.auth, e.org, id, wraps(1, wrap(t, owner, id, 1, owner)))
	require.NoError(t, err)

	// The viewer through a group, the editor by name.
	group := &entity.OrgGroup{ID: gofn.Must(ulid.NewStringULID()), OrgID: e.org, Name: "Ops"}
	_, err = e.db.NewInsert().Model(group).Exec(ctx)
	require.NoError(t, err)
	_, err = e.db.NewInsert().Model(&entity.OrgGroupMember{GroupID: group.ID, OrgID: e.org, UserID: viewer.id()}).
		Exec(ctx)
	require.NoError(t, err)
	e.grant(t, owner, id, userGrant(editor, base.CollectionPermissionEdit),
		&collectiondto.GrantReq{GroupID: group.ID, Permission: base.CollectionPermissionConnect})

	detail, err := e.uc.Get(ctx, owner.auth, e.org, id)
	require.NoError(t, err)
	assert.Equal(t, 2, detail.Data.Collection.Waiting)
	require.Len(t, detail.Data.Grants, 2)
	assert.Equal(t, "ed", detail.Data.Grants[0].Username)
	assert.Equal(t, "Ops", detail.Data.Grants[1].GroupName)
	members := map[string]*collectiondto.MemberResp{}
	for _, m := range detail.Data.Members {
		members[m.Username] = m
	}
	require.Len(t, members, 3)
	assert.True(t, members["olivia"].HasKey)
	assert.Equal(t, base.CollectionPermissionConnect, members["vic"].Permission)
	assert.False(t, members["ed"].HasKey)
	assert.Equal(t, identity.Fingerprint(editor.x25519.PublicKey().Bytes(), editor.ed.Public().(ed25519.PublicKey)),
		members["ed"].Identity.Fingerprint, "to compare before giving it")

	// What the server refuses: a wrap for somebody without access, for another key than theirs, signed by somebody
	// else, or of another version.
	for name, req := range map[string]*collectiondto.WrapsReq{
		"no access": wraps(1, wrap(t, owner, id, 1, stranger)),
		"their key": func() *collectiondto.WrapsReq {
			w := wrap(t, owner, id, 1, stranger)
			w.UserID = editor.id()
			return wraps(1, w)
		}(),
		"signed by another": wraps(1, wrap(t, editor, id, 1, editor)),
	} {
		_, err = e.uc.AddWraps(ctx, owner.auth, e.org, id, req)
		require.ErrorIs(t, err, hperrors.ErrCollectionWrapInvalid, name)
	}
	_, err = e.uc.AddWraps(ctx, owner.auth, e.org, id, wraps(2, wrap(t, owner, id, 2, editor)))
	require.ErrorIs(t, err, hperrors.ErrCollectionKeyVersion)

	_, err = e.uc.AddWraps(ctx, owner.auth, e.org, id, wraps(1, wrap(t, owner, id, 1, editor)))
	require.NoError(t, err)
	_, err = e.uc.AddWraps(ctx, owner.auth, e.org, id, wraps(1, wrap(t, owner, id, 1, editor)))
	require.ErrorIs(t, err, hperrors.ErrCollectionWrapInvalid, "they hold it already")
	assert.Equal(t, 1, e.listed(t, owner, id).Waiting)
	assert.True(t, e.listed(t, editor, id).HasKey)
	assert.Zero(t, e.listed(t, editor, id).Waiting, "not theirs to see")

	// The editor holds the key, and manages nobody's access.
	_, err = e.uc.AddWraps(ctx, editor.auth, e.org, id, wraps(1, wrap(t, editor, id, 1, viewer)))
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	_, err = e.uc.SetGrants(ctx, editor.auth, e.org, id, &collectiondto.SetGrantsReq{})
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	detail, err = e.uc.Get(ctx, editor.auth, e.org, id)
	require.NoError(t, err)
	assert.Empty(t, detail.Data.Members)
	assert.Empty(t, detail.Data.Grants)

	// Granting is to the organization's people and groups, once each.
	_, err = e.uc.SetGrants(ctx, owner.auth, e.org, id, &collectiondto.SetGrantsReq{Grants: []*collectiondto.GrantReq{
		userGrant(editor, base.CollectionPermissionEdit), userGrant(editor, base.CollectionPermissionConnect),
	}})
	require.ErrorIs(t, err, hperrors.ErrValueInvalid)
	_, err = e.uc.SetGrants(ctx, owner.auth, e.org, id, &collectiondto.SetGrantsReq{Grants: []*collectiondto.GrantReq{
		{GroupID: "elsewhere", Permission: base.CollectionPermissionEdit},
	}})
	require.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestSyncServesACollectionToWhoMayUseIt(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.add(t, "olivia", base.OrgRoleOwner)
	admin := e.add(t, "adam", base.OrgRoleAdmin)
	editor := e.add(t, "ed", base.OrgRoleMember)
	viewer := e.add(t, "vic", base.OrgRoleMember)
	stranger := e.add(t, "sam", base.OrgRoleMember)
	id := e.create(t, owner)
	e.grant(t, owner, id, userGrant(editor, base.CollectionPermissionEdit),
		userGrant(viewer, base.CollectionPermissionConnect))

	read := func(p *person) error { return e.sync.CheckAccess(ctx, p.auth, id) }
	write := func(p *person) error {
		_, err := e.sync.DeleteObject(ctx, p.auth, id, "records/x.sw", syncservice.Condition{IfMatch: "x"})
		return err
	}
	require.NoError(t, read(owner))
	require.NoError(t, read(editor))
	require.NoError(t, read(viewer))
	require.ErrorIs(t, read(stranger), hperrors.ErrNotFound)
	require.ErrorIs(t, read(admin), hperrors.ErrNotFound, "an admin manages access, and reads only with a grant")
	require.ErrorIs(t, write(viewer), hperrors.ErrCollectionReadOnly)
	require.ErrorIs(t, write(stranger), hperrors.ErrNotFound)

	// The app lists the collections it syncs after the personal space.
	spaces, err := e.sync.ListSpaces(ctx, viewer.auth)
	require.NoError(t, err)
	require.Len(t, spaces.Data, 2)
	assert.Equal(t, string(base.SpaceTypePersonal), spaces.Data[0].Type)
	assert.Equal(t, id, spaces.Data[1].ID)
	assert.Equal(t, "Production", spaces.Data[1].Name)
	assert.Equal(t, base.CollectionPermissionConnect, spaces.Data[1].Permission)

	// Suspended, they get nothing, and keep their grant.
	setStatus := func(p *person, status base.OrgMemberStatus) {
		_, err := e.db.NewUpdate().Model((*entity.OrgMember)(nil)).Set("status = ?", status).
			Where("org_id = ?", e.org).Where("user_id = ?", p.id()).Exec(ctx)
		require.NoError(t, err)
	}
	setStatus(viewer, base.OrgMemberStatusSuspended)
	require.ErrorIs(t, read(viewer), hperrors.ErrNotFound)
	spaces, err = e.sync.ListSpaces(ctx, viewer.auth)
	require.NoError(t, err)
	assert.Len(t, spaces.Data, 1)
	setStatus(viewer, base.OrgMemberStatusActive)
	require.NoError(t, read(viewer))
}

func TestLosingAccessNeedsANewKeyAndSuspensionDoesNot(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.add(t, "olivia", base.OrgRoleOwner)
	manager := e.add(t, "max", base.OrgRoleMember)
	leaver := e.add(t, "lee", base.OrgRoleMember)
	pauser := e.add(t, "pat", base.OrgRoleMember)
	for _, p := range []*person{owner, manager, leaver, pauser} {
		e.publish(t, p)
	}
	id := e.create(t, owner)
	e.grant(t, owner, id, userGrant(manager, base.CollectionPermissionManage),
		userGrant(leaver, base.CollectionPermissionEdit), userGrant(pauser, base.CollectionPermissionEdit))
	_, err := e.uc.CreateKey(ctx, owner.auth, e.org, id, wraps(1, wrap(t, owner, id, 1, owner),
		wrap(t, owner, id, 1, leaver), wrap(t, owner, id, 1, pauser)))
	require.NoError(t, err)
	assert.Equal(t, 1, e.listed(t, owner, id).Waiting, "the manager")

	// A suspended member keeps the key: nothing to replace.
	_, err = e.db.NewUpdate().Model((*entity.OrgMember)(nil)).Set("status = ?", base.OrgMemberStatusSuspended).
		Where("org_id = ?", e.org).Where("user_id = ?", pauser.id()).Exec(ctx)
	require.NoError(t, err)
	assert.False(t, e.listed(t, owner, id).NeedsNewKey)

	// One whose grant goes held the key: it needs replacing.
	e.grant(t, owner, id, userGrant(manager, base.CollectionPermissionManage),
		userGrant(pauser, base.CollectionPermissionEdit))
	assert.True(t, e.listed(t, owner, id).NeedsNewKey)
	assert.True(t, e.listed(t, manager, id).NeedsNewKey, "a manager sees it")

	// The manager does not hold the current key, so cannot replace it.
	_, err = e.uc.CreateKey(ctx, manager.auth, e.org, id, wraps(2, wrap(t, manager, id, 2, manager),
		wrap(t, manager, id, 2, owner)))
	require.ErrorIs(t, err, hperrors.ErrCollectionNoKey)
	// The owner's app does, for who still has access.
	_, err = e.uc.CreateKey(ctx, owner.auth, e.org, id, wraps(2, wrap(t, owner, id, 2, owner),
		wrap(t, owner, id, 2, manager), wrap(t, owner, id, 2, pauser)))
	require.NoError(t, err)
	got := e.listed(t, owner, id)
	assert.Equal(t, 2, got.KeyVersion)
	assert.False(t, got.NeedsNewKey)

	// The pauser kept both versions: records not sealed again yet still open.
	_, err = e.db.NewUpdate().Model((*entity.OrgMember)(nil)).Set("status = ?", base.OrgMemberStatusActive).
		Where("org_id = ?", e.org).Where("user_id = ?", pauser.id()).Exec(ctx)
	require.NoError(t, err)
	keys, err := e.uc.GetKeys(ctx, pauser.auth, e.org, id)
	require.NoError(t, err)
	require.Len(t, keys.Data.Wraps, 2)
	assert.Equal(t, []int{2, 1}, []int{keys.Data.Wraps[0].Version, keys.Data.Wraps[1].Version})
	_, err = e.uc.GetKeys(ctx, leaver.auth, e.org, id)
	require.ErrorIs(t, err, hperrors.ErrNotFound, "nothing for who lost access")
}
