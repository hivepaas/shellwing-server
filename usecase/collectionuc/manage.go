package collectionuc

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc/collectiondto"
)

func (uc *UC) requireFreeName(ctx context.Context, db database.IDB, orgID, name, exceptID string) error {
	taken, err := uc.collectionRepo.NameTaken(ctx, db, orgID, name, exceptID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if taken {
		return hperrors.NewAlreadyExist("Collection")
	}
	return nil
}

// Create makes a collection, with no key: the first app with Manage on it makes one. Owners and admins make
// collections; one who is not an owner, who have Manage on every collection, is granted Manage on it.
func (uc *UC) Create(ctx context.Context, auth *basedto.Auth, orgID string, req *collectiondto.CollectionReq) (
	*collectiondto.CollectionDataResp, error) {
	now := timeutil.NowUTC()
	collection := &entity.Collection{
		ID:          gofn.Must(ulid.NewStringULID()),
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   auth.UserID(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		// Locked, so that two collections made at once cannot both take a name.
		if _, err := uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
			return hperrors.Wrap(err)
		}
		member, err := uc.memberOf(ctx, tx, auth, orgID)
		if err != nil {
			return err
		}
		if !isAdmin(member) {
			return hperrors.NewForbidden("Collection")
		}
		if err = uc.requireFreeName(ctx, tx, orgID, collection.Name, ""); err != nil {
			return err
		}
		err = uc.spaceRepo.Insert(ctx, tx, &entity.Space{
			ID: collection.ID, Type: base.SpaceTypeCollection, OrgID: orgID, CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.syncRepo.EnsureSpace(ctx, tx, collection.ID); err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.collectionRepo.Insert(ctx, tx, collection); err != nil {
			return hperrors.Wrap(err)
		}
		detail := auditdetail.New().Set("description", collection.Description)
		if member.Role != base.OrgRoleOwner {
			err = uc.collectionRepo.SetGrants(ctx, tx, orgID, collection.ID, []*entity.CollectionGrant{
				newGrant(collection, auth.UserID(), "", base.CollectionPermissionManage, now),
			})
			if err != nil {
				return hperrors.Wrap(err)
			}
			grantees, err := uc.collectionRepo.ListGrantees(ctx, tx, orgID, collection.ID)
			if err != nil {
				return hperrors.Wrap(err)
			}
			detail.Set("grants", describeGrants(grantees))
		}
		return uc.record(ctx, tx, auth, collection, base.AuditLogTypeCollectionUpdate, auditSectionCreate, detail)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := transform(collection, auth.UserID(), base.CollectionPermissionManage, true, nil)
	return &collectiondto.CollectionDataResp{Data: data}, nil
}

func newGrant(collection *entity.Collection, userID, groupID string, permission base.CollectionPermission,
	now time.Time) *entity.CollectionGrant {
	return &entity.CollectionGrant{
		ID:           gofn.Must(ulid.NewStringULID()),
		CollectionID: collection.ID,
		OrgID:        collection.OrgID,
		UserID:       userID,
		GroupID:      groupID,
		Permission:   permission,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Update changes a collection's name and description, for those who manage its access.
func (uc *UC) Update(ctx context.Context, auth *basedto.Auth, orgID, collectionID string,
	req *collectiondto.CollectionReq) (*collectiondto.CollectionDataResp, error) {
	var c *caller
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if _, err := uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
			return hperrors.Wrap(err)
		}
		var err error
		if c, err = uc.reach(ctx, tx, auth, orgID, collectionID, true); err != nil {
			return err
		}
		if !c.managesAccess() {
			return hperrors.NewForbidden("Collection")
		}
		if err = uc.requireFreeName(ctx, tx, orgID, req.Name, collectionID); err != nil {
			return err
		}
		detail := auditdetail.New().
			Set("before", map[string]string{"name": c.collection.Name, "description": c.collection.Description})
		c.collection.Name, c.collection.Description = req.Name, req.Description
		if err = uc.collectionRepo.Update(ctx, tx, c.collection); err != nil {
			return hperrors.Wrap(err)
		}
		detail.Set("description", req.Description)
		return uc.record(ctx, tx, auth, c.collection, base.AuditLogTypeCollectionUpdate, auditSectionUpdate, detail)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := transform(c.collection, auth.UserID(), c.permission, true, nil)
	return &collectiondto.CollectionDataResp{Data: data}, nil
}

// Delete removes a collection, for owners and admins. Its files go after the store's trash delay; nobody syncs it from
// now on.
func (uc *UC) Delete(ctx context.Context, auth *basedto.Auth, orgID, collectionID string) (
	*collectiondto.DeleteCollectionResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		c, err := uc.reach(ctx, tx, auth, orgID, collectionID, true)
		if err != nil {
			return err
		}
		if !isAdmin(c.member) {
			return hperrors.NewForbidden("Collection")
		}
		if err = uc.collectionRepo.Delete(ctx, tx, orgID, collectionID); err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.spaceRepo.Delete(ctx, tx, collectionID); err != nil {
			return hperrors.Wrap(err)
		}
		deleteAfter := timeutil.NowUTC().Add(config.Current().Sync.TrashDelay)
		if err = uc.syncRepo.TrashAll(ctx, tx, collectionID, deleteAfter); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.record(ctx, tx, auth, c.collection, base.AuditLogTypeCollectionUpdate, auditSectionDelete,
			auditdetail.New())
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &collectiondto.DeleteCollectionResp{}, nil
}

// SetGrants makes these the collection's grants, for those who manage its access. A person granted is a member of the
// organization who is not removed; a group is one of its groups. Those granted wait for the key until an app holding
// it gives it to them.
func (uc *UC) SetGrants(ctx context.Context, auth *basedto.Auth, orgID, collectionID string,
	req *collectiondto.SetGrantsReq) (*collectiondto.SetGrantsResp, error) {
	var grantees []*entity.CollectionGrantee
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		c, err := uc.reach(ctx, tx, auth, orgID, collectionID, true)
		if err != nil {
			return err
		}
		if !c.managesAccess() {
			return hperrors.NewForbidden("Collection")
		}
		before, err := uc.collectionRepo.ListGrantees(ctx, tx, orgID, collectionID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		grants, err := uc.checkGrants(ctx, tx, c.collection, req.Grants)
		if err != nil {
			return err
		}
		if err = uc.collectionRepo.SetGrants(ctx, tx, orgID, collectionID, grants); err != nil {
			return hperrors.Wrap(err)
		}
		if grantees, err = uc.collectionRepo.ListGrantees(ctx, tx, orgID, collectionID); err != nil {
			return hperrors.Wrap(err)
		}
		detail := auditdetail.New().Set("before", describeGrants(before)).Set("grants", describeGrants(grantees))
		return uc.record(ctx, tx, auth, c.collection, base.AuditLogTypeCollectionUpdate, auditSectionGrants, detail)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := make([]*collectiondto.GrantResp, 0, len(grantees))
	for _, g := range grantees {
		data = append(data, collectiondto.TransformGrant(g))
	}
	return &collectiondto.SetGrantsResp{Data: data}, nil
}

// Grantable returns whom the collection can be granted to, for those who manage its access: in the web app, and in
// the app, whose token reaches no other list of the organization's people.
func (uc *UC) Grantable(ctx context.Context, auth *basedto.Auth, orgID, collectionID string) (
	*collectiondto.GetGrantableResp, error) {
	c, err := uc.reach(ctx, uc.db, auth, orgID, collectionID, false)
	if err != nil {
		return nil, err
	}
	if !c.managesAccess() {
		return nil, hperrors.NewForbidden("Collection")
	}
	members, err := uc.orgRepo.ListMembers(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	groups, err := uc.groupRepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := &collectiondto.GrantableResp{
		People: make([]*collectiondto.GrantablePersonResp, 0, len(members)),
		Groups: make([]*collectiondto.GrantableGroupResp, 0, len(groups)),
	}
	for _, m := range members {
		if m.Status == base.OrgMemberStatusRemoved {
			continue
		}
		data.People = append(data.People, &collectiondto.GrantablePersonResp{
			UserID:    m.UserID,
			Username:  m.Username,
			FullName:  m.FullName,
			Role:      m.Role,
			Suspended: m.Status == base.OrgMemberStatusSuspended,
		})
	}
	for _, g := range groups {
		data.Groups = append(data.Groups, &collectiondto.GrantableGroupResp{GroupID: g.ID, Name: g.Name})
	}
	return &collectiondto.GetGrantableResp{Data: data}, nil
}

func (uc *UC) checkGrants(ctx context.Context, db database.IDB, collection *entity.Collection,
	reqs []*collectiondto.GrantReq) ([]*entity.CollectionGrant, error) {
	now := timeutil.NowUTC()
	seen := map[string]bool{}
	grants := make([]*entity.CollectionGrant, 0, len(reqs))
	for _, req := range reqs {
		key := "user:" + req.UserID
		if req.GroupID != "" {
			key = "group:" + req.GroupID
		}
		if seen[key] {
			return nil, hperrors.Wrap(hperrors.ErrValueInvalid).WithParam("Name", key)
		}
		seen[key] = true
		if req.UserID != "" {
			member, err := uc.orgRepo.GetMember(ctx, db, collection.OrgID, req.UserID)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if member == nil || member.Status == base.OrgMemberStatusRemoved {
				return nil, hperrors.NewNotFound("Member")
			}
		} else {
			group, err := uc.groupRepo.Get(ctx, db, collection.OrgID, req.GroupID)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if group == nil {
				return nil, hperrors.NewNotFound("Group")
			}
		}
		grants = append(grants, newGrant(collection, req.UserID, req.GroupID, req.Permission, now))
	}
	return grants, nil
}

// describeGrants names each grant for the audit log: "user:alice=edit", "group:Ops=connect".
func describeGrants(grantees []*entity.CollectionGrantee) []string {
	out := make([]string, 0, len(grantees))
	for _, g := range grantees {
		if g.GroupID != "" {
			out = append(out, fmt.Sprintf("group:%s=%s", g.GroupName, g.Permission))
		} else {
			out = append(out, fmt.Sprintf("user:%s=%s", g.Username, g.Permission))
		}
	}
	return out
}
