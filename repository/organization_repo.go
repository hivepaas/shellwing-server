package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type OrganizationRepo interface {
	// GetBySlug returns nil when no organization has the slug.
	GetBySlug(ctx context.Context, db database.IDB, slug string) (*entity.Organization, error)
	// InsertIfAbsent makes the organization unless its slug is taken: two instances starting together make one.
	InsertIfAbsent(ctx context.Context, db database.IDB, org *entity.Organization) error

	// AddMembers adds the members who are not there yet, and leaves alone those who are, removed ones included.
	AddMembers(ctx context.Context, db database.IDB, members []*entity.OrgMember) error
	// UpdateMember saves a member's role, status and principals.
	UpdateMember(ctx context.Context, db database.IDB, member *entity.OrgMember) error
	DeleteMember(ctx context.Context, db database.IDB, orgID, userID string) error
	// AddAllUsers makes every user who is not a member of the organization an active one.
	AddAllUsers(ctx context.Context, db database.IDB, orgID string) error
	// CountActiveOwners counts the active owners whose accounts are active, but for the user named.
	CountActiveOwners(ctx context.Context, db database.IDB, orgID, exceptUserID string) (int, error)

	// LockByID returns the organization with its row locked, or nil. What changes its CAs takes this first.
	LockByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error)
	// GetByID returns nil when no organization has the id.
	GetByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error)
	SetSSHCertTTL(ctx context.Context, db database.IDB, id string, minutes int) error
	// SetSSHCertExtensions stores what the organization's certificates are allowed to do.
	SetSSHCertExtensions(ctx context.Context, db database.IDB, id string, extensions []string) error
	// SetSSHKRLToken stores the hash a target server fetches the revocation list with, or clears it with an empty
	// hash.
	SetSSHKRLToken(ctx context.Context, db database.IDB, id, tokenHash string, now time.Time) error
	// GetBySSHKRLToken returns the organization that token belongs to, or nil. The token names it by itself, so a
	// caller holding one needs no session.
	GetBySSHKRLToken(ctx context.Context, db database.IDB, tokenHash string) (*entity.Organization, error)
	// GetMember returns nil when the user is not a member.
	GetMember(ctx context.Context, db database.IDB, orgID, userID string) (*entity.OrgMember, error)
	// ListMembers returns the organization's members who are users still, removed ones included, by username.
	ListMembers(ctx context.Context, db database.IDB, orgID string) ([]*entity.OrgMemberUser, error)
	// ListMembershipsOfUser returns the user's memberships, as active members, of organizations that are active.
	ListMembershipsOfUser(ctx context.Context, db database.IDB, userID string) ([]*entity.OrgMember, error)
	SetSSHPrincipals(ctx context.Context, db database.IDB, orgID, userID string, principals []string) error
}

type organizationRepo struct {
}

func NewOrganizationRepo() OrganizationRepo {
	return &organizationRepo{}
}

func (repo *organizationRepo) GetBySlug(ctx context.Context, db database.IDB,
	slug string) (*entity.Organization, error) {
	org := &entity.Organization{}
	err := db.NewSelect().Model(org).Where("organization.slug = ?", slug).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return org, nil
}

func (repo *organizationRepo) InsertIfAbsent(ctx context.Context, db database.IDB, org *entity.Organization) error {
	_, err := db.NewInsert().Model(org).On("CONFLICT (slug) WHERE deleted_at IS NULL DO NOTHING").Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) AddMembers(ctx context.Context, db database.IDB, members []*entity.OrgMember) error {
	if len(members) == 0 {
		return nil
	}
	_, err := db.NewInsert().Model(&members).On("CONFLICT (org_id, user_id) DO NOTHING").Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) UpdateMember(ctx context.Context, db database.IDB, member *entity.OrgMember) error {
	_, err := db.NewUpdate().Model((*entity.OrgMember)(nil)).
		Set("role = ?", member.Role).
		Set("status = ?", member.Status).
		Set("ssh_principals = ?", pgdialect.Array(orEmpty(member.SSHPrincipals))).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("org_id = ?", member.OrgID).Where("user_id = ?", member.UserID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) CountActiveOwners(ctx context.Context, db database.IDB,
	orgID, exceptUserID string) (int, error) {
	count, err := db.NewSelect().TableExpr("org_members AS m").
		Join("JOIN users AS u ON u.id = m.user_id AND u.deleted_at IS NULL AND u.status = ?", base.UserStatusActive).
		Where("m.org_id = ?", orgID).
		Where("m.role = ?", base.OrgRoleOwner).
		Where("m.status = ?", base.OrgMemberStatusActive).
		Where("m.user_id <> ?", exceptUserID).
		Count(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int(count), nil
}

func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}

func (repo *organizationRepo) DeleteMember(ctx context.Context, db database.IDB, orgID, userID string) error {
	_, err := db.NewDelete().Model((*entity.OrgMember)(nil)).
		Where("org_id = ?", orgID).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) AddAllUsers(ctx context.Context, db database.IDB, orgID string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO org_members (org_id, user_id, role, status)
		SELECT ?, u.id, 'member', 'active'
		FROM users u
		WHERE u.deleted_at IS NULL
		ON CONFLICT (org_id, user_id) DO NOTHING`, orgID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) getByID(ctx context.Context, db database.IDB, id string,
	forUpdate bool) (*entity.Organization, error) {
	org := &entity.Organization{}
	query := db.NewSelect().Model(org).Where("organization.id = ?", id)
	if forUpdate {
		query = query.For("UPDATE")
	}
	err := query.Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return org, nil
}

func (repo *organizationRepo) LockByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error) {
	return repo.getByID(ctx, db, id, true)
}

func (repo *organizationRepo) GetByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error) {
	return repo.getByID(ctx, db, id, false)
}

func (repo *organizationRepo) SetSSHCertTTL(ctx context.Context, db database.IDB, id string, minutes int) error {
	_, err := db.NewUpdate().Model((*entity.Organization)(nil)).
		Set("ssh_cert_ttl_minutes = ?", minutes).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) SetSSHCertExtensions(ctx context.Context, db database.IDB, id string,
	extensions []string) error {
	if extensions == nil {
		extensions = []string{}
	}
	_, err := db.NewUpdate().Model((*entity.Organization)(nil)).
		Set("ssh_cert_extensions = ?", pgdialect.Array(extensions)).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) SetSSHKRLToken(ctx context.Context, db database.IDB, id, tokenHash string,
	now time.Time) error {
	q := db.NewUpdate().Model((*entity.Organization)(nil)).Set("updated_at = CURRENT_TIMESTAMP").Where("id = ?", id)
	if tokenHash == "" {
		q = q.Set("ssh_krl_token_hash = NULL").Set("ssh_krl_token_created_at = NULL")
	} else {
		q = q.Set("ssh_krl_token_hash = ?", tokenHash).Set("ssh_krl_token_created_at = ?", now)
	}
	if _, err := q.Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *organizationRepo) GetBySSHKRLToken(ctx context.Context, db database.IDB, tokenHash string) (
	*entity.Organization, error) {
	if tokenHash == "" {
		return nil, nil
	}
	org := &entity.Organization{}
	err := db.NewSelect().Model(org).Where("ssh_krl_token_hash = ?", tokenHash).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return org, nil
}

func (repo *organizationRepo) GetMember(ctx context.Context, db database.IDB,
	orgID, userID string) (*entity.OrgMember, error) {
	member := &entity.OrgMember{}
	err := db.NewSelect().Model(member).Where("org_id = ?", orgID).Where("user_id = ?", userID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return member, nil
}

func (repo *organizationRepo) ListMembershipsOfUser(ctx context.Context, db database.IDB,
	userID string) ([]*entity.OrgMember, error) {
	var members []*entity.OrgMember
	err := db.NewSelect().Model(&members).
		Join("JOIN organizations AS o ON o.id = org_member.org_id AND o.deleted_at IS NULL AND o.status = ?",
			base.OrgStatusActive).
		Where("org_member.user_id = ?", userID).
		Where("org_member.status = ?", base.OrgMemberStatusActive).
		Order("org_member.created_at").
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return members, nil
}

func (repo *organizationRepo) ListMembers(ctx context.Context, db database.IDB,
	orgID string) ([]*entity.OrgMemberUser, error) {
	members := []*entity.OrgMemberUser{}
	err := db.NewSelect().TableExpr("org_members AS m").
		Join("JOIN users AS u ON u.id = m.user_id AND u.deleted_at IS NULL").
		ColumnExpr("m.user_id, m.role, m.status, m.ssh_principals, u.username, u.email, u.full_name").
		ColumnExpr("u.status AS account_status").
		Where("m.org_id = ?", orgID).
		OrderExpr("u.username").
		Scan(ctx, &members)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return members, nil
}

func (repo *organizationRepo) SetSSHPrincipals(ctx context.Context, db database.IDB, orgID, userID string,
	principals []string) error {
	_, err := db.NewUpdate().Model((*entity.OrgMember)(nil)).
		Set("ssh_principals = ?", pgdialect.Array(principals)).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("org_id = ?", orgID).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
