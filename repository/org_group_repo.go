package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// OrgGroupRepo stores an organization's groups and who is in them. Every call names the organization: one's group is
// never reached through another's.
type OrgGroupRepo interface {
	// List returns the organization's groups, by name.
	List(ctx context.Context, db database.IDB, orgID string) ([]*entity.OrgGroup, error)
	// Get returns nil when the organization has no such group.
	Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.OrgGroup, error)
	// NameTaken says whether another group of the organization has the name, whatever its case.
	NameTaken(ctx context.Context, db database.IDB, orgID, name, exceptID string) (bool, error)
	Insert(ctx context.Context, db database.IDB, group *entity.OrgGroup) error
	// Update saves a group's name, description and principals.
	Update(ctx context.Context, db database.IDB, group *entity.OrgGroup) error
	Delete(ctx context.Context, db database.IDB, orgID, id string) error

	// ListMembers returns who is in the organization's groups, by group name then username.
	ListMembers(ctx context.Context, db database.IDB, orgID string) ([]*entity.OrgGroupMemberUser, error)
	// SetMembers makes these users the group's members, and nobody else.
	SetMembers(ctx context.Context, db database.IDB, orgID, groupID string, userIDs []string) error
	// RemoveFromAll takes a member out of every group of the organization.
	RemoveFromAll(ctx context.Context, db database.IDB, orgID, userID string) error
	// PrincipalsOf returns the accounts the member's groups grant, sorted, once each.
	PrincipalsOf(ctx context.Context, db database.IDB, orgID, userID string) ([]string, error)
}

type orgGroupRepo struct {
}

func NewOrgGroupRepo() OrgGroupRepo {
	return &orgGroupRepo{}
}

func (repo *orgGroupRepo) List(ctx context.Context, db database.IDB, orgID string) ([]*entity.OrgGroup, error) {
	groups := []*entity.OrgGroup{}
	err := db.NewSelect().Model(&groups).Where("org_id = ?", orgID).OrderExpr("lower(name)").Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return groups, nil
}

func (repo *orgGroupRepo) Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.OrgGroup, error) {
	group := &entity.OrgGroup{}
	err := db.NewSelect().Model(group).Where("org_id = ?", orgID).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return group, nil
}

func (repo *orgGroupRepo) NameTaken(ctx context.Context, db database.IDB, orgID, name, exceptID string) (bool, error) {
	exists, err := db.NewSelect().Model((*entity.OrgGroup)(nil)).
		Where("org_id = ?", orgID).
		Where("lower(name) = lower(?)", name).
		Where("id <> ?", exceptID).
		Exists(ctx)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return exists, nil
}

func (repo *orgGroupRepo) Insert(ctx context.Context, db database.IDB, group *entity.OrgGroup) error {
	if _, err := db.NewInsert().Model(group).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *orgGroupRepo) Update(ctx context.Context, db database.IDB, group *entity.OrgGroup) error {
	_, err := db.NewUpdate().Model((*entity.OrgGroup)(nil)).
		Set("name = ?", group.Name).
		Set("description = ?", group.Description).
		Set("ssh_principals = ?", pgdialect.Array(orEmpty(group.SSHPrincipals))).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("org_id = ?", group.OrgID).Where("id = ?", group.ID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *orgGroupRepo) Delete(ctx context.Context, db database.IDB, orgID, id string) error {
	_, err := db.NewDelete().Model((*entity.OrgGroup)(nil)).Where("org_id = ?", orgID).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *orgGroupRepo) ListMembers(ctx context.Context, db database.IDB,
	orgID string) ([]*entity.OrgGroupMemberUser, error) {
	members := []*entity.OrgGroupMemberUser{}
	err := db.NewSelect().TableExpr("org_group_members AS gm").
		Join("JOIN org_groups AS g ON g.id = gm.group_id").
		Join("JOIN users AS u ON u.id = gm.user_id AND u.deleted_at IS NULL").
		ColumnExpr("gm.group_id, g.name AS group_name, gm.user_id, u.username, u.full_name").
		Where("gm.org_id = ?", orgID).
		OrderExpr("lower(g.name), u.username").
		Scan(ctx, &members)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return members, nil
}

func (repo *orgGroupRepo) SetMembers(ctx context.Context, db database.IDB, orgID, groupID string,
	userIDs []string) error {
	_, err := db.NewDelete().Model((*entity.OrgGroupMember)(nil)).
		Where("org_id = ?", orgID).Where("group_id = ?", groupID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(userIDs) == 0 {
		return nil
	}
	members := make([]*entity.OrgGroupMember, 0, len(userIDs))
	for _, userID := range userIDs {
		members = append(members, &entity.OrgGroupMember{GroupID: groupID, OrgID: orgID, UserID: userID})
	}
	if _, err = db.NewInsert().Model(&members).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *orgGroupRepo) RemoveFromAll(ctx context.Context, db database.IDB, orgID, userID string) error {
	_, err := db.NewDelete().Model((*entity.OrgGroupMember)(nil)).
		Where("org_id = ?", orgID).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *orgGroupRepo) PrincipalsOf(ctx context.Context, db database.IDB, orgID, userID string) ([]string, error) {
	principals := []string{}
	err := db.NewSelect().TableExpr("org_group_members AS gm").
		Join("JOIN org_groups AS g ON g.id = gm.group_id").
		ColumnExpr("DISTINCT unnest(g.ssh_principals) AS principal").
		Where("gm.org_id = ?", orgID).Where("gm.user_id = ?", userID).
		OrderExpr("principal").
		Scan(ctx, &principals)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return principals, nil
}
