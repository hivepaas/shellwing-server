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
		// A new group has no members, so this cannot grant its maker anything; it goes through noSelfGrant all the
		// same, so that every path which touches principals is covered by the one rule.
		err := uc.noSelfGrant(ctx, tx, auth, orgID, func() error {
			return hperrors.Wrap(uc.groupRepo.Insert(ctx, tx, group))
		})
		if err != nil {
			return err
		}
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupCreate, auditdetail.New().
			Set("name", group.Name).Set("principals", principals))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: orgdto.TransformGroup(group, nil)}, nil
}

// UpdateGroup changes a group's name, description and SSH accounts. Its members' certificates are revoked and their
// next ones follow the new accounts. An admin cannot grant themselves through a group they are in - see noSelfGrant.
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
		changedAccounts := !slices.Equal(group.SSHPrincipals, principals)
		group.Name, group.Description, group.SSHPrincipals = req.Name, req.Description, principals
		err = uc.noSelfGrant(ctx, tx, auth, orgID, func() error {
			return hperrors.Wrap(uc.groupRepo.Update(ctx, tx, group))
		})
		if err != nil {
			return err
		}
		// A rename alone reaches nobody's certificates; a change to the accounts reaches every member's.
		var revoked int64
		if changedAccounts {
			if revoked, err = uc.revokeOfGroup(ctx, tx, orgID, groupID); err != nil {
				return err
			}
		}
		members, err := uc.groupRepo.ListMembers(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data = orgdto.TransformGroup(group, members)
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupUpdate, auditdetail.New().
			Set("name", group.Name).Set("principals", principals).Set("certsRevoked", revoked))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: data}, nil
}

// DeleteGroup deletes a group. Its members' certificates are revoked, and they lose what it granted.
func (uc *UC) DeleteGroup(ctx context.Context, auth *basedto.Auth, orgID, groupID string) (
	*orgdto.DeleteGroupResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		group, err := uc.lockGroup(ctx, tx, auth, orgID, groupID)
		if err != nil {
			return err
		}
		// Before the group goes: who is in it is what the rows about to be deleted say.
		revoked, err := uc.revokeOfGroup(ctx, tx, orgID, groupID)
		if err != nil {
			return err
		}
		if err = uc.groupRepo.Delete(ctx, tx, orgID, groupID); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupDelete,
			auditdetail.New().Set("name", group.Name).Set("certsRevoked", revoked))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.DeleteGroupResp{}, nil
}

// SetGroupMembers makes these members the group's, and nobody else. An admin cannot add themselves to a group that
// grants accounts they do not have - see noSelfGrant. Each is a member of the organization, suspended
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
		// Who was in it, read before the change: somebody taken out keeps their certificates otherwise, and what the
		// group granted is in them.
		were, err := uc.groupRepo.MemberIDsOf(ctx, tx, orgID, groupID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = uc.noSelfGrant(ctx, tx, auth, orgID, func() error {
			return hperrors.Wrap(uc.groupRepo.SetMembers(ctx, tx, orgID, groupID, userIDs))
		})
		if err != nil {
			return err
		}
		reaches := slices.Clone(were)
		reaches = append(reaches, userIDs...)
		slices.Sort(reaches)
		revoked, err := uc.sshCertRepo.RevokeOfUsers(ctx, tx, orgID, slices.Compact(reaches), timeutil.NowUTC())
		if err != nil {
			return hperrors.Wrap(err)
		}
		members, err := uc.groupRepo.ListMembers(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data = orgdto.TransformGroup(group, members)
		return uc.recordGroupChange(ctx, tx, auth, group, auditSectionGroupMembers, auditdetail.New().
			Set("name", group.Name).Set("userIds", userIDs).Set("certsRevoked", revoked))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.GroupDataResp{Data: data}, nil
}

// revokeOfGroup refuses the certificates of everybody in the group, for a change to what it grants or to the group
// itself. What each of them may log in as has changed, and a certificate they hold still says what it did.
func (uc *UC) revokeOfGroup(ctx context.Context, tx database.Tx, orgID, groupID string) (int64, error) {
	userIDs, err := uc.groupRepo.MemberIDsOf(ctx, tx, orgID, groupID)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	revoked, err := uc.sshCertRepo.RevokeOfUsers(ctx, tx, orgID, userIDs, timeutil.NowUTC())
	return revoked, hperrors.Wrap(err)
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
