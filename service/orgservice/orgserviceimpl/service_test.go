package orgserviceimpl

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
)

// These need a real PostgreSQL with the migrations applied: point SHW_TEST_PG_DSN at one to enable them.
func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("SHW_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SHW_TEST_PG_DSN not set, skipping organization test")
	}
	conf, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	bunDB := bun.NewDB(stdlib.OpenDB(*conf), pgdialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })
	return &database.DB{DB: bunDB}
}

func insertUser(t *testing.T, db *database.DB, role base.UserRole, status base.UserStatus) *entity.User {
	t.Helper()
	id := gofn.Must(ulid.NewStringULID())
	user := &entity.User{ID: id, Username: id, Role: role, Status: status,
		SecurityOption: base.UserSecurityPasswordOnly}
	_, err := db.NewInsert().Model(user).Exec(context.Background())
	require.NoError(t, err)
	return user
}

func memberOf(t *testing.T, db *database.DB, orgID, userID string) *entity.OrgMember {
	t.Helper()
	member := &entity.OrgMember{}
	err := db.NewSelect().Model(member).Where("org_id = ?", orgID).Where("user_id = ?", userID).Scan(context.Background())
	if err != nil {
		return nil
	}
	return member
}

func TestDefaultOrganization(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	s := New(repository.NewOrganizationRepo())

	before := insertUser(t, db, base.UserRoleAdmin, base.UserStatusPending)
	org, err := s.EnsureDefaultOrg(ctx, db)
	require.NoError(t, err)
	again, err := s.EnsureDefaultOrg(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, org.ID, again.ID, "one organization, however often")

	// Users from before it are made members, whatever they are on the server.
	member := memberOf(t, db, org.ID, before.ID)
	require.NotNil(t, member)
	assert.Equal(t, base.OrgRoleMember, member.Role)
	assert.Equal(t, base.OrgMemberStatusActive, member.Status)

	// A new one, then the changes a user goes through: none is the organization's.
	user := insertUser(t, db, base.UserRoleMember, base.UserStatusActive)
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{user}))
	member = memberOf(t, db, org.ID, user.ID)
	require.NotNil(t, member)
	assert.Equal(t, base.OrgRoleMember, member.Role)
	assert.Equal(t, base.OrgMemberStatusActive, member.Status)

	user.Role, user.Status = base.UserRoleAdmin, base.UserStatusDisabled
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{user}))
	member = memberOf(t, db, org.ID, user.ID)
	assert.Equal(t, base.OrgRoleMember, member.Role, "an admin of the server is not one of the organization")
	assert.Equal(t, base.OrgMemberStatusActive, member.Status)

	// The first admin owns it, and stays the owner when their user is saved.
	require.NoError(t, s.MakeOwner(ctx, db, user))
	assert.Equal(t, base.OrgRoleOwner, memberOf(t, db, org.ID, user.ID).Role)
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{user}))
	assert.Equal(t, base.OrgRoleOwner, memberOf(t, db, org.ID, user.ID).Role)

	// A member an admin removed is not made one again by saving their user.
	removed := insertUser(t, db, base.UserRoleMember, base.UserStatusActive)
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{removed}))
	_, err = db.NewUpdate().Model((*entity.OrgMember)(nil)).Set("status = ?", base.OrgMemberStatusRemoved).
		Where("org_id = ?", org.ID).Where("user_id = ?", removed.ID).Exec(ctx)
	require.NoError(t, err)
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{removed}))
	assert.Equal(t, base.OrgMemberStatusRemoved, memberOf(t, db, org.ID, removed.ID).Status)

	// Deleted, either way, they are no member.
	user.DeletedAt = timeutil.NowUTC()
	require.NoError(t, s.SyncUsers(ctx, db, []*entity.User{user}))
	assert.Nil(t, memberOf(t, db, org.ID, user.ID))
	require.NoError(t, s.RemoveUser(ctx, db, before.ID))
	assert.Nil(t, memberOf(t, db, org.ID, before.ID))
}
