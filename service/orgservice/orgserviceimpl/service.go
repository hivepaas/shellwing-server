package orgserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/orgservice"
)

const defaultOrgName = "Default"

func New(orgRepo repository.OrganizationRepo) orgservice.Service {
	return &service{orgRepo: orgRepo}
}

type service struct {
	orgRepo repository.OrganizationRepo
}

func (s *service) EnsureDefaultOrg(ctx context.Context, db database.IDB) (*entity.Organization, error) {
	org, err := s.defaultOrg(ctx, db)
	if err != nil {
		return nil, err
	}
	if org == nil {
		timeNow := timeutil.NowUTC()
		err = s.orgRepo.InsertIfAbsent(ctx, db, &entity.Organization{
			ID:        gofn.Must(ulid.NewStringULID()),
			Slug:      base.DefaultOrgSlug,
			Name:      defaultOrgName,
			Status:    base.OrgStatusActive,
			UpdateVer: 1,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if org, err = s.defaultOrg(ctx, db); err != nil {
			return nil, err
		}
		if org == nil {
			return nil, hperrors.NewNotFound("Default organization")
		}
	}
	if err = s.orgRepo.AddAllUsers(ctx, db, org.ID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return org, nil
}

func (s *service) MakeOwner(ctx context.Context, db database.IDB, user *entity.User) error {
	org, err := s.requireDefaultOrg(ctx, db)
	if err != nil {
		return err
	}
	member := newMember(org.ID, user)
	member.Role = base.OrgRoleOwner
	return s.orgRepo.UpsertMembers(ctx, db, []*entity.OrgMember{member}) //nolint:wrapcheck // the repo's own
}

func (s *service) SyncUsers(ctx context.Context, db database.IDB, users []*entity.User) error {
	if len(users) == 0 {
		return nil
	}
	org, err := s.requireDefaultOrg(ctx, db)
	if err != nil {
		return err
	}
	members := make([]*entity.OrgMember, 0, len(users))
	for _, user := range users {
		if !user.DeletedAt.IsZero() {
			if err = s.orgRepo.DeleteMember(ctx, db, org.ID, user.ID); err != nil {
				return hperrors.Wrap(err)
			}
			continue
		}
		members = append(members, newMember(org.ID, user))
	}
	return s.orgRepo.UpsertMembers(ctx, db, members) //nolint:wrapcheck // the repo's own
}

func (s *service) RemoveUser(ctx context.Context, db database.IDB, userID string) error {
	org, err := s.requireDefaultOrg(ctx, db)
	if err != nil {
		return err
	}
	return s.orgRepo.DeleteMember(ctx, db, org.ID, userID) //nolint:wrapcheck // the repo's own
}

func (s *service) defaultOrg(ctx context.Context, db database.IDB) (*entity.Organization, error) {
	org, err := s.orgRepo.GetBySlug(ctx, db, base.DefaultOrgSlug)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return org, nil
}

func (s *service) requireDefaultOrg(ctx context.Context, db database.IDB) (*entity.Organization, error) {
	org, err := s.defaultOrg(ctx, db)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, hperrors.NewNotFound("Default organization")
	}
	return org, nil
}

func newMember(orgID string, user *entity.User) *entity.OrgMember {
	timeNow := timeutil.NowUTC()
	role := base.OrgRoleMember
	if user.IsAdmin() {
		role = base.OrgRoleAdmin
	}
	return &entity.OrgMember{
		OrgID:     orgID,
		UserID:    user.ID,
		Role:      role,
		Status:    memberStatusOf(user.Status),
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
}

func memberStatusOf(status base.UserStatus) base.OrgMemberStatus {
	switch status {
	case base.UserStatusActive:
		return base.OrgMemberStatusActive
	case base.UserStatusPending:
		return base.OrgMemberStatusInvited
	case base.UserStatusDisabled:
		return base.OrgMemberStatusDisabled
	}
	return base.OrgMemberStatusDisabled
}
