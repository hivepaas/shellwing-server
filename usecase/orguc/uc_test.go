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
	for _, table := range []string{"organizations", "org_members", "ssh_cas", "users"} {
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
	e := &env{uc: New(db, repository.NewOrganizationRepo(), repository.NewSSHCARepo(), audit), db: db, audit: audit}
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

	for _, status := range []base.OrgMemberStatus{base.OrgMemberStatusDisabled, base.OrgMemberStatusInvited} {
		person := e.addMember(t, e.org, base.OrgRoleMember, status)
		_, err = e.uc.SetSSHPrincipals(ctx, admin, e.org, person.UserID(),
			&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
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
		OrgID: other, UserID: person.UserID(), Role: base.OrgRoleMember, Status: base.OrgMemberStatusDisabled,
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
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusInvited)
	_ = e.addMember(t, e.addOrg(t), base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err := e.uc.SetSSHPrincipals(ctx, admin, e.org, member.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
	require.NoError(t, err)

	resp, err := e.uc.ListMembers(ctx, admin, e.org)
	require.NoError(t, err)
	got := map[string]*orgdto.MemberResp{}
	for _, m := range resp.Data {
		got[m.UserID] = m
	}
	require.Len(t, got, 2, "only this organization's")
	assert.Equal(t, []string{"deploy"}, got[member.UserID()].SSHPrincipals)
	assert.Equal(t, base.OrgMemberStatusInvited, got[member.UserID()].Status)
	assert.Equal(t, []string{}, got[admin.UserID()].SSHPrincipals, "none until granted")
	assert.NotEmpty(t, got[admin.UserID()].Username)

	_, err = e.uc.ListMembers(ctx, member, e.org)
	require.ErrorIs(t, err, hperrors.ErrNotFound, "an invited member is no member yet")
	other := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err = e.uc.ListMembers(ctx, other, e.org)
	require.ErrorIs(t, err, hperrors.ErrForbidden)
}
