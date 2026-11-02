package repository

import (
	"context"
	"database/sql"
	"errors"

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

	// UpsertMembers adds members and updates those there. An owner stays an owner: nothing else takes that from them.
	UpsertMembers(ctx context.Context, db database.IDB, members []*entity.OrgMember) error
	DeleteMember(ctx context.Context, db database.IDB, orgID, userID string) error
	// AddAllUsers makes every user who is not a member of the organization one, as their role and status say.
	AddAllUsers(ctx context.Context, db database.IDB, orgID string) error

	// LockByID returns the organization with its row locked, or nil. What changes its CAs takes this first.
	LockByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error)
	// GetByID returns nil when no organization has the id.
	GetByID(ctx context.Context, db database.IDB, id string) (*entity.Organization, error)
	SetSSHCertTTL(ctx context.Context, db database.IDB, id string, minutes int) error
	// GetMember returns nil when the user is not a member.
	GetMember(ctx context.Context, db database.IDB, orgID, userID string) (*entity.OrgMember, error)
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

func (repo *organizationRepo) UpsertMembers(ctx context.Context, db database.IDB, members []*entity.OrgMember) error {
	if len(members) == 0 {
		return nil
	}
	_, err := db.NewInsert().Model(&members).
		On("CONFLICT (org_id, user_id) DO UPDATE").
		Set("role = CASE WHEN org_member.role = 'owner' THEN org_member.role ELSE EXCLUDED.role END").
		Set("status = EXCLUDED.status").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
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
		SELECT ?, u.id,
		       CASE WHEN u.role = 'admin' THEN 'admin' ELSE 'member' END,
		       CASE u.status WHEN 'pending' THEN 'invited' ELSE u.status END
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
