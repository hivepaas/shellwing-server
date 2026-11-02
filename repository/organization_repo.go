package repository

import (
	"context"
	"database/sql"
	"errors"

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
