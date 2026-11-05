package orguc

import (
	"context"
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

const (
	auditSectionGroupCreate  = "create"
	auditSectionGroupUpdate  = "update"
	auditSectionGroupDelete  = "delete"
	auditSectionGroupMembers = "members"
)

// ListGroups returns the organization's groups with their members, for its admins.
func (uc *UC) ListGroups(ctx context.Context, auth *basedto.Auth, orgID string) (*orgdto.ListGroupsResp, error) {
	if err := uc.requireAdmin(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	groups, err := uc.groupRepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	members, err := uc.groupRepo.ListMembers(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := make([]*orgdto.GroupResp, 0, len(groups))
	for _, g := range groups {
		data = append(data, orgdto.TransformGroup(g, members))
	}
	return &orgdto.ListGroupsResp{Data: data}, nil
}

// CreateGroup makes a group with no members. Its name is unique in the organization, whatever its case.
func (uc *UC) CreateGroup(ctx context.Context, auth *basedto.Auth, orgID string, req *orgdto.GroupReq) (
	*orgdto.GroupDataResp, error) {
	principals, err := cleanPrincipals(req.SSHPrincipals)
	if err != nil {
		return nil, err
	}
	now := timeutil.NowUTC()
	group := &entity.OrgGroup{
		ID:            gofn.Must(ulid.NewStringULID()),
		OrgID:         orgID,
		Name:          req.Name,
		Description:   req.Description,
		SSHPrincipals: principals,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.lockAsAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		if err := uc.requireFreeName(ctx, tx, orgID, group.Name, ""); err != nil {
			return err
		}
		if err := uc.groupRepo.Insert(ctx, tx, group); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupCreate, auditdetail.New().
			Set("name", group.Name).Set("principals", principals))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: orgdto.TransformGroup(group, nil)}, nil
}

// UpdateGroup changes a group's name, description and SSH accounts. Its members' next certificates follow.
func (uc *UC) UpdateGroup(ctx context.Context, auth *basedto.Auth, orgID, groupID string, req *orgdto.GroupReq) (
	*orgdto.GroupDataResp, error) {
	principals, err := cleanPrincipals(req.SSHPrincipals)
	if err != nil {
		return nil, err
	}
	var data *orgdto.GroupResp
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		group, err := uc.lockGroup(ctx, tx, auth, orgID, groupID)
		if err != nil {
			return err
		}
		if err = uc.requireFreeName(ctx, tx, orgID, req.Name, groupID); err != nil {
			return err
		}
		group.Name, group.Description, group.SSHPrincipals = req.Name, req.Description, principals
		if err = uc.groupRepo.Update(ctx, tx, group); err != nil {
			return hperrors.Wrap(err)
		}
		members, err := uc.groupRepo.ListMembers(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data = orgdto.TransformGroup(group, members)
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupUpdate, auditdetail.New().
			Set("name", group.Name).Set("principals", principals))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: data}, nil
}

// DeleteGroup deletes a group. Its members lose what it granted, from their next certificate on.
func (uc *UC) DeleteGroup(ctx context.Context, auth *basedto.Auth, orgID, groupID string) (
	*orgdto.DeleteGroupResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		group, err := uc.lockGroup(ctx, tx, auth, orgID, groupID)
		if err != nil {
			return err
		}
		if err = uc.groupRepo.Delete(ctx, tx, orgID, groupID); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupDelete,
			auditdetail.New().Set("name", group.Name))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.DeleteGroupResp{}, nil
}

// SetGroupMembers makes these members the group's, and nobody else. Each is a member of the organization, suspended
// ones included; a removed one is not.
func (uc *UC) SetGroupMembers(ctx context.Context, auth *basedto.Auth, orgID, groupID string,
	req *orgdto.SetGroupMembersReq) (*orgdto.GroupDataResp, error) {
	userIDs := slices.Clone(req.UserIDs)
	slices.Sort(userIDs)
	userIDs = slices.Compact(userIDs)
	var data *orgdto.GroupResp
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		group, err := uc.lockGroup(ctx, tx, auth, orgID, groupID)
		if err != nil {
			return err
		}
		for _, userID := range userIDs {
			member, err := uc.orgRepo.GetMember(ctx, tx, orgID, userID)
			if err != nil {
				return hperrors.Wrap(err)
			}
			if member == nil || member.Status == base.OrgMemberStatusRemoved {
				return hperrors.NewNotFound("Member")
			}
		}
		if err = uc.groupRepo.SetMembers(ctx, tx, orgID, groupID, userIDs); err != nil {
			return hperrors.Wrap(err)
		}
		members, err := uc.groupRepo.ListMembers(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data = orgdto.TransformGroup(group, members)
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupMembers, auditdetail.New().
			Set("name", group.Name).Set("userIds", userIDs))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: data}, nil
}

// lockAsAdmin checks the person is an admin, with the organization locked: two groups made at once cannot both take a
// name.
func (uc *UC) lockAsAdmin(ctx context.Context, tx database.Tx, auth *basedto.Auth, orgID string) error {
	if _, err := uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
		return hperrors.Wrap(err)
	}
	return uc.requireAdmin(ctx, tx, auth, orgID)
}

// lockGroup returns the organization's group for an admin to change, with the organization locked.
func (uc *UC) lockGroup(ctx context.Context, tx database.Tx, auth *basedto.Auth, orgID, groupID string) (
	*entity.OrgGroup, error) {
	if err := uc.lockAsAdmin(ctx, tx, auth, orgID); err != nil {
		return nil, err
	}
	group, err := uc.groupRepo.Get(ctx, tx, orgID, groupID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if group == nil {
		return nil, hperrors.NewNotFound("Group")
	}
	return group, nil
}

func (uc *UC) requireFreeName(ctx context.Context, db database.IDB, orgID, name, exceptID string) error {
	taken, err := uc.groupRepo.NameTaken(ctx, db, orgID, name, exceptID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if taken {
		return hperrors.NewAlreadyExist("Group")
	}
	return nil
}

func (uc *UC) recordGroupChange(ctx context.Context, db database.IDB, auth *basedto.Auth, group *entity.OrgGroup,
	section string, detail *auditdetail.Builder) error {
	return uc.record(ctx, db, auth, &auditservice.Entry{
		Type:    base.AuditLogTypeOrgGroupUpdate,
		Section: section,
		ResType: base.ResourceTypeOrgGroup,
		ResID:   group.ID,
		Detail:  detail.Set("orgId", group.OrgID).String(),
	})
}
