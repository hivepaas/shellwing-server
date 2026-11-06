package orguc

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

// fakeAudit keeps entries, and fails on demand: a certificate with no record must not be handed out.
type fakeAudit struct {
	auditservice.Service
	entries []*auditservice.Entry
	fail    error
}

func (f *fakeAudit) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	if f.fail != nil {
		return f.fail
	}
	f.entries = append(f.entries, entry)
	return nil
}

type env struct {
	uc    *UC
	db    *database.DB
	audit *fakeAudit
	org   string
}

// These need a real PostgreSQL with the migrations applied. Point SHW_TEST_PG_DSN at one to enable them. Each test has
// a schema of its own.
func newEnv(t *testing.T) *env {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping organization test")
	}
	key, err := datakey.Generate()
	require.NoError(t, err)
	datakey.SetActive(key)

	ctx := context.Background()
	schema := "org_" + strings.ToLower(gofn.Must(ulid.NewStringULID()))
	open := func(conf *pgx.ConnConfig) *bun.DB { return bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New()) }
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	admin := open(conf)
	t.Cleanup(func() { _ = admin.Close() })
	stmts := []string{"CREATE SCHEMA " + schema}
	for _, table := range []string{"organizations", "org_members", "ssh_cas", "users", "org_groups",
		"org_group_members", "collection_grants"} {
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
	e := &env{uc: New(db, repository.NewOrganizationRepo(), repository.NewSSHCARepo(), repository.NewOrgGroupRepo(),
		repository.NewCollectionRepo(), audit), db: db, audit: audit}
	e.org = e.addOrg(t)
	return e
}

func (e *env) addOrg(t *testing.T) string {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	_, err := e.db.NewInsert().Model(&entity.Organization{
		ID: id, Slug: strings.ToLower(id), Name: "Org", Status: base.OrgStatusActive, UpdateVer: 1,
	}).Exec(context.Background())
	require.NoError(t, err)
	return id
}

func (e *env) addMember(t *testing.T, orgID string, role base.OrgRole, status base.OrgMemberStatus) *basedto.Auth {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	_, err := e.db.NewInsert().Model(&entity.User{
		ID: id, Username: "user-" + strings.ToLower(id), Role: base.UserRoleMember, Status: base.UserStatusActive,
		SecurityOption: base.UserSecurityPasswordOnly,
	}).Exec(context.Background())
	require.NoError(t, err)
	_, err = e.db.NewInsert().Model(&entity.OrgMember{
		OrgID: orgID, UserID: id, Role: role, Status: status,
	}).Exec(context.Background())
	require.NoError(t, err)
	return &basedto.Auth{User: &basedto.User{
		User:       &entity.User{ID: id, Username: "user-" + id},
		AuthClaims: &jwtsession.AuthClaims{UID: "device-1", UserID: id, Scope: jwtsession.ScopeSync},
	}}
}

func newKey(t *testing.T) string {
	t.Helper()
	ca, err := sshca.GenerateCA()
	require.NoError(t, err)
	return ca.PublicKey
}

func issueReq(t *testing.T) *orgdto.IssueSSHCertificateReq {
	t.Helper()
	return &orgdto.IssueSSHCertificateReq{PublicKey: newKey(t)}
}

func TestACertificateNamesWhatAnAdminGranted(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)

	_, err := e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals, "nobody has an account until granted one")

	_, err = e.uc.SetSSHPrincipals(ctx, admin, e.org, member.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy", "app", "deploy"}})
	require.NoError(t, err)

	_, err = e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHCANotSetUp)

	created, err := e.uc.CreateSSHCA(ctx, admin, e.org)
	require.NoError(t, err)

	resp, err := e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.NoError(t, err)
	assert.Equal(t, []string{"app", "deploy"}, resp.Data.Principals)
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(resp.Data.Certificate)) //nolint:dogsled // only the key
	require.NoError(t, err)
	cert, ok := key.(*ssh.Certificate)
	require.True(t, ok)
	assert.Equal(t, []string{"app", "deploy"}, cert.ValidPrincipals)
	assert.Contains(t, cert.KeyId, "device=device-1")
	assert.Contains(t, cert.KeyId, member.UserID())
	assert.Equal(t, created.Data.PublicKey,
		strings.TrimSpace(string(ssh.MarshalAuthorizedKey(cert.SignatureKey))))
	assert.InDelta(t, 480*time.Minute, time.Until(resp.Data.ValidBefore), float64(time.Minute))

	_, err = e.uc.UpdateSSHSettings(ctx, admin, e.org, &orgdto.UpdateSSHSettingsReq{CertTTLMinutes: 10})
	require.NoError(t, err)
	resp, err = e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.NoError(t, err)
	assert.InDelta(t, 10*time.Minute, time.Until(resp.Data.ValidBefore), float64(time.Minute))

	_, err = e.uc.SetSSHPrincipals(ctx, admin, e.org, member.UserID(), &orgdto.SetSSHPrincipalsReq{})
	require.NoError(t, err)
	_, err = e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals, "taken away")
}

func TestAnOwnersOwnCertificateNeedsAGrantToo(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org)
	require.NoError(t, err)
	_, err = e.uc.IssueSSHCertificate(ctx, owner, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals)
}

func TestOnlyAnActiveMemberGetsOne(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, admin, e.org)
	require.NoError(t, err)

	for _, status := range []base.OrgMemberStatus{base.OrgMemberStatusSuspended, base.OrgMemberStatusRemoved} {
		person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
		_, err = e.uc.SetSSHPrincipals(ctx, admin, e.org, person.UserID(),
			&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
		require.NoError(t, err)
		_, err = e.db.NewUpdate().Model((*entity.OrgMember)(nil)).Set("status = ?", status).
			Where("org_id = ?", e.org).Where("user_id = ?", person.UserID()).Exec(ctx)
		require.NoError(t, err)
		_, err = e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
		require.ErrorIs(t, err, hperrors.ErrNotFound, string(status))
	}
}

func TestOneOrganizationReachesNothingOfAnother(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	ownerA := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	caA, err := e.uc.CreateSSHCA(ctx, ownerA, e.org)
	require.NoError(t, err)

	orgB := e.addOrg(t)
	ownerB := e.addMember(t, orgB, base.OrgRoleOwner, base.OrgMemberStatusActive)
	memberA := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)

	_, err = e.uc.ListSSHCAs(ctx, ownerB, e.org)
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.CreateSSHCA(ctx, ownerB, e.org)
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.DeleteSSHCA(ctx, ownerB, e.org, caA.Data.ID)
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.IssueSSHCertificate(ctx, ownerB, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.SetSSHPrincipals(ctx, ownerB, e.org, memberA.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrNotFound)

	// Its own admin naming the other's CA or member finds neither.
	_, err = e.uc.DeleteSSHCA(ctx, ownerB, orgB, caA.Data.ID)
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.SetSSHPrincipals(ctx, ownerB, orgB, memberA.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrNotFound)
}

func TestAMemberCannotDoWhatAnAdminDoes(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)

	_, err := e.uc.CreateSSHCA(ctx, member, e.org)
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	_, err = e.uc.SetSSHPrincipals(ctx, member, e.org, member.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrForbidden, "nobody grants themselves an account")
	_, err = e.uc.GetSSHPrincipals(ctx, member, e.org, member.UserID())
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	_, err = e.uc.UpdateSSHSettings(ctx, member, e.org, &orgdto.UpdateSSHSettingsReq{CertTTLMinutes: 1440})
	require.ErrorIs(t, err, hperrors.ErrForbidden)
}

func TestANewCARetiresTheOldAndDestroysItsKey(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	first, err := e.uc.CreateSSHCA(ctx, admin, e.org)
	require.NoError(t, err)

	_, err = e.uc.DeleteSSHCA(ctx, admin, e.org, first.Data.ID)
	require.ErrorIs(t, err, hperrors.ErrSSHCAActive)

	second, err := e.uc.CreateSSHCA(ctx, admin, e.org)
	require.NoError(t, err)

	list, err := e.uc.ListSSHCAs(ctx, admin, e.org)
	require.NoError(t, err)
	require.Len(t, list.Data, 2)
	assert.Equal(t, second.Data.ID, list.Data[0].ID)
	assert.Equal(t, base.SSHCAStatusActive, list.Data[0].Status)
	assert.Equal(t, base.SSHCAStatusRetired, list.Data[1].Status)

	var private *string
	require.NoError(t, e.db.NewRaw("SELECT private_key FROM ssh_cas WHERE id = ?", first.Data.ID).Scan(ctx, &private))
	assert.Nil(t, private, "a retired CA can sign nothing")
	require.NoError(t, e.db.NewRaw("SELECT private_key FROM ssh_cas WHERE id = ?", second.Data.ID).Scan(ctx, &private))
	require.NotNil(t, private)
	assert.True(t, strings.HasPrefix(*private, "hpenc:"), "sealed, not in the clear")

	_, err = e.uc.DeleteSSHCA(ctx, admin, e.org, first.Data.ID)
	require.NoError(t, err)
	list, err = e.uc.ListSSHCAs(ctx, admin, e.org)
	require.NoError(t, err)
	assert.Len(t, list.Data, 1)
}

func TestTwoNewCAsAtOnceLeaveOneActive(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)

	var group errgroup.Group
	for range 8 {
		group.Go(func() error {
			_, err := e.uc.CreateSSHCA(ctx, admin, e.org)
			return err
		})
	}
	require.NoError(t, group.Wait())

	list, err := e.uc.ListSSHCAs(ctx, admin, e.org)
	require.NoError(t, err)
	require.Len(t, list.Data, 8)
	active := 0
	for _, ca := range list.Data {
		if ca.Status == base.SSHCAStatusActive {
			active++
		}
	}
	assert.Equal(t, 1, active)
}

func TestACertificateNobodyRecordedIsNotHandedOut(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, admin, e.org)
	require.NoError(t, err)
	_, err = e.uc.SetSSHPrincipals(ctx, admin, e.org, admin.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
	require.NoError(t, err)

	req := issueReq(t)
	e.audit.fail = assert.AnError
	_, err = e.uc.IssueSSHCertificate(ctx, admin, e.org, req)
	require.Error(t, err)
	e.audit.fail = nil

	e.audit.entries = nil
	resp, err := e.uc.IssueSSHCertificate(ctx, admin, e.org, req)
	require.NoError(t, err)
	require.Len(t, e.audit.entries, 1)
	entry := e.audit.entries[0]
	assert.Equal(t, base.AuditLogTypeSSHCertIssue, entry.Type)
	assert.Contains(t, entry.Detail, "deploy")
	assert.Contains(t, entry.Detail, "device-1")
	assert.NotContains(t, entry.Detail, resp.Data.Certificate)
}

func TestListOrgsLeavesOutWhereThePersonIsNotActive(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	other := e.addOrg(t)
	_, err := e.db.NewInsert().Model(&entity.OrgMember{
		OrgID: other, UserID: person.UserID(), Role: base.OrgRoleMember, Status: base.OrgMemberStatusSuspended,
	}).Exec(ctx)
	require.NoError(t, err)

	resp, err := e.uc.ListOrgs(ctx, person)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, e.org, resp.Data[0].ID)
	assert.Equal(t, 480, resp.Data[0].SSHCertTTLMinutes)
	assert.Empty(t, resp.Data[0].SSHPrincipals)
}

func TestAnAdminListsTheMembersWithWhatTheyMayLogInAs(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	suspended := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusSuspended)
	_ = e.addMember(t, e.addOrg(t), base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err := e.uc.SetSSHPrincipals(ctx, admin, e.org, suspended.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
	require.NoError(t, err)

	resp, err := e.uc.ListMembers(ctx, admin, e.org)
	require.NoError(t, err)
	got := map[string]*orgdto.MemberResp{}
	for _, m := range resp.Data {
		got[m.UserID] = m
	}
	require.Len(t, got, 2, "only this organization's")
	assert.Equal(t, []string{"deploy"}, got[suspended.UserID()].SSHPrincipals)
	assert.Equal(t, base.OrgMemberStatusSuspended, got[suspended.UserID()].Status)
	assert.Equal(t, base.UserStatusActive, got[suspended.UserID()].AccountStatus)
	assert.Equal(t, []string{}, got[admin.UserID()].SSHPrincipals, "none until granted")
	assert.NotEmpty(t, got[admin.UserID()].Username)

	_, err = e.uc.ListMembers(ctx, suspended, e.org)
	require.ErrorIs(t, err, hperrors.ErrNotFound, "a suspended member is no member for now")
}

func TestAMemberSeesTheOthersByNameAndRoleOnly(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_ = e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusSuspended)
	_, err := e.uc.SetSSHPrincipals(ctx, admin, e.org, admin.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"root"}})
	require.NoError(t, err)

	resp, err := e.uc.ListMembers(ctx, member, e.org)
	require.NoError(t, err)
	require.Len(t, resp.Data, 2, "the active ones")
	for _, m := range resp.Data {
		assert.NotEmpty(t, m.Username)
		assert.NotEmpty(t, m.Role)
		assert.Empty(t, m.Email)
		assert.Empty(t, m.Status)
		assert.Nil(t, m.SSHPrincipals, "what others may log in as is for admins")
	}
}

func TestWhoChangesWhichRole(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	role := func(by *basedto.Auth, who *basedto.Auth, to base.OrgRole) error {
		_, err := e.uc.UpdateMemberRole(ctx, by, e.org, who.UserID(), &orgdto.UpdateMemberRoleReq{Role: to})
		return err
	}
	roleOf := func(who *basedto.Auth) base.OrgRole {
		m, err := repository.NewOrganizationRepo().GetMember(ctx, e.db, e.org, who.UserID())
		require.NoError(t, err)
		return m.Role
	}

	require.ErrorIs(t, role(member, member, base.OrgRoleAdmin), hperrors.ErrForbidden)
	require.ErrorIs(t, role(admin, member, base.OrgRoleOwner), hperrors.ErrForbidden, "only an owner makes one")
	require.ErrorIs(t, role(admin, owner, base.OrgRoleMember), hperrors.ErrForbidden, "nor changes one")

	require.NoError(t, role(admin, member, base.OrgRoleAdmin))
	assert.Equal(t, base.OrgRoleAdmin, roleOf(member))
	require.NoError(t, role(admin, member, base.OrgRoleMember), "an admin changes another admin")
	assert.Equal(t, base.OrgRoleMember, roleOf(member))

	require.NoError(t, role(owner, member, base.OrgRoleOwner))
	assert.Equal(t, base.OrgRoleOwner, roleOf(member))
	require.NoError(t, role(member, owner, base.OrgRoleAdmin), "the new owner hands the old one admin")
	assert.Equal(t, base.OrgRoleAdmin, roleOf(owner))

	require.Len(t, e.audit.entries, 4)
	last := e.audit.entries[3]
	assert.Equal(t, base.AuditLogTypeOrgMemberUpdate, last.Type)
	assert.Equal(t, "role", last.Section)
	assert.Equal(t, owner.UserID(), last.ResID)
	assert.Contains(t, last.Detail, `"from":"owner"`)
	assert.Contains(t, last.Detail, `"to":"admin"`)
}

func TestTheLastActiveOwnerStaysOne(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	other := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.db.NewUpdate().Model((*entity.User)(nil)).Set("status = ?", base.UserStatusDisabled).
		Where("id = ?", other.UserID()).Exec(ctx)
	require.NoError(t, err)
	_ = e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusSuspended)

	// Neither a disabled account nor a suspended owner keeps the organization.
	_, err = e.uc.UpdateMemberRole(ctx, owner, e.org, owner.UserID(),
		&orgdto.UpdateMemberRoleReq{Role: base.OrgRoleAdmin})
	require.ErrorIs(t, err, hperrors.ErrOrgLastOwner)
	for _, status := range []base.OrgMemberStatus{base.OrgMemberStatusSuspended, base.OrgMemberStatusRemoved} {
		_, err = e.uc.UpdateMemberStatus(ctx, owner, e.org, owner.UserID(),
			&orgdto.UpdateMemberStatusReq{Status: status})
		require.ErrorIs(t, err, hperrors.ErrOrgLastOwner, string(status))
	}

	// With another active owner, they may leave.
	next := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err = e.uc.UpdateMemberRole(ctx, owner, e.org, next.UserID(), &orgdto.UpdateMemberRoleReq{Role: base.OrgRoleOwner})
	require.NoError(t, err)
	_, err = e.uc.UpdateMemberStatus(ctx, owner, e.org, owner.UserID(),
		&orgdto.UpdateMemberStatusReq{Status: base.OrgMemberStatusRemoved})
	require.NoError(t, err)
}

func TestARemovedMemberLosesWhatTheyWereGranted(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org)
	require.NoError(t, err)
	person := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	_, err = e.uc.SetSSHPrincipals(ctx, owner, e.org, person.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
	require.NoError(t, err)
	status := func(to base.OrgMemberStatus) error {
		_, err := e.uc.UpdateMemberStatus(ctx, owner, e.org, person.UserID(), &orgdto.UpdateMemberStatusReq{Status: to})
		return err
	}

	// Suspended, they get nothing, and keep what they had.
	require.NoError(t, status(base.OrgMemberStatusSuspended))
	_, err = e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	require.NoError(t, status(base.OrgMemberStatusActive))
	_, err = e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)

	// Removed, it is gone; brought back, they are a member with nothing.
	_, err = e.db.NewInsert().Model(&entity.CollectionGrant{
		ID: "g1", CollectionID: "c1", OrgID: e.org, UserID: person.UserID(), Permission: base.CollectionPermissionEdit,
	}).Exec(ctx)
	require.NoError(t, err)
	require.NoError(t, status(base.OrgMemberStatusRemoved))
	grants, err := e.db.NewSelect().Model((*entity.CollectionGrant)(nil)).Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, grants, "their collections' grants")
	_, err = e.uc.GetSSHPrincipals(ctx, owner, e.org, person.UserID())
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.SetSSHPrincipals(ctx, owner, e.org, person.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = e.uc.UpdateMemberRole(ctx, owner, e.org, person.UserID(),
		&orgdto.UpdateMemberRoleReq{Role: base.OrgRoleAdmin})
	require.ErrorIs(t, err, hperrors.ErrNotFound)
	require.NoError(t, status(base.OrgMemberStatusActive))
	resp, err := e.uc.ListMembers(ctx, owner, e.org)
	require.NoError(t, err)
	for _, m := range resp.Data {
		if m.UserID == person.UserID() {
			assert.Equal(t, base.OrgRoleMember, m.Role)
			assert.Empty(t, m.SSHPrincipals)
		}
	}
	_, err = e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals)
}

func TestAMemberMayLeaveAndDoNothingElse(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	_ = e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	other := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	status := func(by, who *basedto.Auth, to base.OrgMemberStatus) error {
		_, err := e.uc.UpdateMemberStatus(ctx, by, e.org, who.UserID(), &orgdto.UpdateMemberStatusReq{Status: to})
		return err
	}

	require.ErrorIs(t, status(member, other, base.OrgMemberStatusRemoved), hperrors.ErrForbidden)
	require.ErrorIs(t, status(member, member, base.OrgMemberStatusSuspended), hperrors.ErrForbidden)
	require.NoError(t, status(member, member, base.OrgMemberStatusRemoved))
	_, err := e.uc.ListMembers(ctx, member, e.org)
	require.ErrorIs(t, err, hperrors.ErrNotFound, "gone")

	// An admin takes out another admin; nobody but an owner takes out an owner.
	require.NoError(t, status(admin, other, base.OrgMemberStatusSuspended))
	second := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	require.NoError(t, status(admin, second, base.OrgMemberStatusRemoved))
}

func grant(t *testing.T, e *env, admin, who *basedto.Auth, principals ...string) {
	t.Helper()
	_, err := e.uc.SetSSHPrincipals(context.Background(), admin, e.org, who.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: principals})
	require.NoError(t, err)
}

func newGroup(t *testing.T, e *env, admin *basedto.Auth, name string, principals ...string) string {
	t.Helper()
	resp, err := e.uc.CreateGroup(context.Background(), admin, e.org,
		&orgdto.GroupReq{Name: name, SSHPrincipals: principals})
	require.NoError(t, err)
	return resp.Data.ID
}

func setMembers(e *env, admin *basedto.Auth, groupID string, who ...*basedto.Auth) error {
	ids := make([]string, 0, len(who))
	for _, w := range who {
		ids = append(ids, w.UserID())
	}
	_, err := e.uc.SetGroupMembers(context.Background(), admin, e.org, groupID, &orgdto.SetGroupMembersReq{UserIDs: ids})
	return err
}

func certifiedFor(t *testing.T, e *env, who *basedto.Auth) []string {
	t.Helper()
	resp, err := e.uc.IssueSSHCertificate(context.Background(), who, e.org, issueReq(t))
	require.NoError(t, err)
	return resp.Data.Principals
}

func TestAGroupsAccountsAreEachMembers(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org)
	require.NoError(t, err)
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	grant(t, e, owner, person, "app")
	ops := newGroup(t, e, owner, "Ops", "root", "deploy", "app")
	require.NoError(t, setMembers(e, owner, ops, person))

	assert.Equal(t, []string{"app", "deploy", "root"}, certifiedFor(t, e, person), "theirs and the group's, once each")
	orgs, err := e.uc.ListOrgs(ctx, person)
	require.NoError(t, err)
	assert.Equal(t, []string{"app", "deploy", "root"}, orgs.Data[0].SSHPrincipals)
	members, err := e.uc.ListMembers(ctx, owner, e.org)
	require.NoError(t, err)
	for _, m := range members.Data {
		if m.UserID == person.UserID() {
			assert.Equal(t, []string{"app"}, m.SSHPrincipals, "their own")
			require.Len(t, m.Groups, 1)
			assert.Equal(t, "Ops", m.Groups[0].Name)
		}
	}

	_, err = e.uc.UpdateGroup(ctx, owner, e.org, ops, &orgdto.GroupReq{Name: "Ops", SSHPrincipals: []string{"deploy"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"app", "deploy"}, certifiedFor(t, e, person))
	_, err = e.uc.DeleteGroup(ctx, owner, e.org, ops)
	require.NoError(t, err)
	assert.Equal(t, []string{"app"}, certifiedFor(t, e, person))

	// A group alone is enough.
	other := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err = e.uc.IssueSSHCertificate(ctx, other, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals)
	require.NoError(t, setMembers(e, owner, newGroup(t, e, owner, "Web", "web"), other))
	assert.Equal(t, []string{"web"}, certifiedFor(t, e, other))
}

func TestAGroupNameIsTakenWhateverItsCase(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_ = newGroup(t, e, owner, "Ops")
	_, err := e.uc.CreateGroup(ctx, owner, e.org, &orgdto.GroupReq{Name: "ops"})
	require.ErrorIs(t, err, hperrors.ErrAlreadyExist)
	web := newGroup(t, e, owner, "Web")
	_, err = e.uc.UpdateGroup(ctx, owner, e.org, web, &orgdto.GroupReq{Name: "OPS"})
	require.ErrorIs(t, err, hperrors.ErrAlreadyExist)
	_, err = e.uc.UpdateGroup(ctx, owner, e.org, web, &orgdto.GroupReq{Name: "web", Description: "The web team"})
	require.NoError(t, err, "its own name in another case")
	_, err = e.uc.CreateGroup(ctx, owner, e.org, &orgdto.GroupReq{Name: "Bad", SSHPrincipals: []string{"a b"}})
	require.ErrorIs(t, err, hperrors.ErrSSHPrincipalInvalid)

	// Another organization has its own names.
	e.org = e.addOrg(t)
	other := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_ = newGroup(t, e, other, "Ops")
}

func TestOnlyAdminsTouchGroups(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	ops := newGroup(t, e, owner, "Ops", "root")

	_, err := e.uc.ListGroups(ctx, member, e.org)
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	_, err = e.uc.CreateGroup(ctx, member, e.org, &orgdto.GroupReq{Name: "Mine", SSHPrincipals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	_, err = e.uc.UpdateGroup(ctx, member, e.org, ops, &orgdto.GroupReq{Name: "Ops", SSHPrincipals: []string{"root"}})
	require.ErrorIs(t, err, hperrors.ErrForbidden)
	require.ErrorIs(t, setMembers(e, member, ops, member), hperrors.ErrForbidden, "nobody joins root by themselves")
	_, err = e.uc.DeleteGroup(ctx, member, e.org, ops)
	require.ErrorIs(t, err, hperrors.ErrForbidden)
}

func TestAGroupHoldsOnlyItsOrganizationsMembers(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	ops := newGroup(t, e, owner, "Ops", "root")
	orgA := e.org

	orgB := e.addOrg(t)
	outsider := e.addMember(t, orgB, base.OrgRoleOwner, base.OrgMemberStatusActive)
	require.ErrorIs(t, setMembers(e, owner, ops, outsider), hperrors.ErrNotFound)
	removed := e.addMember(t, orgA, base.OrgRoleMember, base.OrgMemberStatusRemoved)
	require.ErrorIs(t, setMembers(e, owner, ops, removed), hperrors.ErrNotFound)
	suspended := e.addMember(t, orgA, base.OrgRoleMember, base.OrgMemberStatusSuspended)
	require.NoError(t, setMembers(e, owner, ops, suspended), "suspended is a member still")

	// The other organization's owner reaches none of it, through either organization.
	for _, org := range []string{orgA, orgB} {
		_, err := e.uc.UpdateGroup(ctx, outsider, org, ops, &orgdto.GroupReq{Name: "Mine", SSHPrincipals: []string{"x"}})
		require.ErrorIs(t, err, hperrors.ErrNotFound, org)
		_, err = e.uc.DeleteGroup(ctx, outsider, org, ops)
		require.ErrorIs(t, err, hperrors.ErrNotFound, org)
	}
	groups, err := e.uc.ListGroups(ctx, outsider, orgB)
	require.NoError(t, err)
	assert.Empty(t, groups.Data)
}

func TestARemovedMemberLeavesEveryGroup(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org)
	require.NoError(t, err)
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	ops := newGroup(t, e, owner, "Ops", "root")
	require.NoError(t, setMembers(e, owner, ops, person, owner))

	status := func(to base.OrgMemberStatus) {
		_, err := e.uc.UpdateMemberStatus(ctx, owner, e.org, person.UserID(), &orgdto.UpdateMemberStatusReq{Status: to})
		require.NoError(t, err)
	}
	status(base.OrgMemberStatusSuspended)
	status(base.OrgMemberStatusActive)
	assert.Equal(t, []string{"root"}, certifiedFor(t, e, person), "suspending keeps the groups")

	status(base.OrgMemberStatusRemoved)
	status(base.OrgMemberStatusActive)
	_, err = e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals)
	groups, err := e.uc.ListGroups(ctx, owner, e.org)
	require.NoError(t, err)
	require.Len(t, groups.Data[0].Members, 1, "the others stay")
	assert.Equal(t, owner.UserID(), groups.Data[0].Members[0].UserID)
}
