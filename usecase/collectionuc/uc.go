// Package collectionuc is an organization's collections: who may do what in them, and their keys as each person's app
// wrapped them (docs/design/collections.md in the app's repository). The server never holds a key that opens one.
package collectionuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc/collectiondto"
	"github.com/hivepaas/shellwing-server/usecase/identityuc/identitydto"
)

const (
	auditSectionCreate = "create"
	auditSectionUpdate = "update"
	auditSectionDelete = "delete"
	auditSectionGrants = "grants"
	auditSectionKey    = "key"
	auditSectionWrap   = "wrap"
)

type UC struct {
	db             *database.DB
	orgRepo        repository.OrganizationRepo
	groupRepo      repository.OrgGroupRepo
	collectionRepo repository.CollectionRepo
	spaceRepo      repository.SpaceRepo
	syncRepo       repository.SyncRepo
	identityRepo   repository.UserIdentityRepo
	auditService   auditservice.Service
}

func New(
	db *database.DB,
	orgRepo repository.OrganizationRepo,
	groupRepo repository.OrgGroupRepo,
	collectionRepo repository.CollectionRepo,
	spaceRepo repository.SpaceRepo,
	syncRepo repository.SyncRepo,
	identityRepo repository.UserIdentityRepo,
	auditService auditservice.Service,
) *UC {
	return &UC{
		db:             db,
		orgRepo:        orgRepo,
		groupRepo:      groupRepo,
		collectionRepo: collectionRepo,
		spaceRepo:      spaceRepo,
		syncRepo:       syncRepo,
		identityRepo:   identityRepo,
		auditService:   auditService,
	}
}

// memberOf is the person's membership of an active organization. An organization they are not an active member of is
// simply not found: saying more would tell a caller which ids exist.
func (uc *UC) memberOf(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID string) (
	*entity.OrgMember, error) {
	org, err := uc.orgRepo.GetByID(ctx, db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if org == nil || org.Status != base.OrgStatusActive {
		return nil, hperrors.NewNotFound("Organization")
	}
	member, err := uc.orgRepo.GetMember(ctx, db, orgID, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if member == nil || member.Status != base.OrgMemberStatusActive {
		return nil, hperrors.NewNotFound("Organization")
	}
	return member, nil
}

func isAdmin(member *entity.OrgMember) bool {
	return member.Role == base.OrgRoleOwner || member.Role == base.OrgRoleAdmin
}

// caller is a person reaching one collection: their membership and what they may do in it.
type caller struct {
	member     *entity.OrgMember
	collection *entity.Collection
	permission base.CollectionPermission
}

// managesAccess says whether they may grant others and give them the key: an admin, or somebody with Manage.
func (c *caller) managesAccess() bool {
	return isAdmin(c.member) || c.permission == base.CollectionPermissionManage
}

// reach returns the collection for the person, locked when lock is set. A collection they may neither use nor manage
// is not found.
func (uc *UC) reach(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID, collectionID string,
	lock bool) (*caller, error) {
	member, err := uc.memberOf(ctx, db, auth, orgID)
	if err != nil {
		return nil, err
	}
	get := uc.collectionRepo.Get
	if lock {
		get = uc.collectionRepo.Lock
	}
	collection, err := get(ctx, db, orgID, collectionID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if collection == nil {
		return nil, hperrors.NewNotFound("Collection")
	}
	entitlements, err := uc.collectionRepo.Entitlements(ctx, db, orgID, collectionID, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	c := &caller{member: member, collection: collection}
	if len(entitlements) > 0 && entitlements[0].Served() {
		c.permission = entitlements[0].Permission
	}
	if c.permission == "" && !isAdmin(member) {
		return nil, hperrors.NewNotFound("Collection")
	}
	return c, nil
}

func (uc *UC) record(ctx context.Context, db database.IDB, auth *basedto.Auth, collection *entity.Collection,
	logType base.AuditLogType, section string, detail *auditdetail.Builder) error {
	return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Auth:    auth,
		Type:    logType,
		Section: section,
		ResType: base.ResourceTypeCollection,
		ResID:   collection.ID,
		Detail:  detail.Set("orgId", collection.OrgID).Set("name", collection.Name).String(),
	}))
}

// standing is what the server knows of one collection's access and key, for those who manage it.
type standing struct {
	entitled map[string]*entity.CollectionEntitlement
	holders  map[string]bool
}

// waiting counts those served who do not hold the current key.
func (s *standing) waiting() int {
	n := 0
	for userID, e := range s.entitled {
		if e.Served() && !s.holders[userID] {
			n++
		}
	}
	return n
}

// needsNewKey says somebody holding the current key has no access any more: not suspended, which keeps it, but no
// longer granted, no longer an owner, or removed.
func (s *standing) needsNewKey() bool {
	for userID := range s.holders {
		if s.entitled[userID] == nil {
			return true
		}
	}
	return false
}

// standings gathers, for each collection of the organization, who has access and who holds its key.
func (uc *UC) standings(ctx context.Context, db database.IDB, orgID, collectionID string) (
	map[string]*standing, error) {
	entitlements, err := uc.collectionRepo.Entitlements(ctx, db, orgID, collectionID, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	holders, err := uc.collectionRepo.KeyHolders(ctx, db, orgID, collectionID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := map[string]*standing{}
	of := func(id string) *standing {
		s := out[id]
		if s == nil {
			s = &standing{entitled: map[string]*entity.CollectionEntitlement{}, holders: map[string]bool{}}
			out[id] = s
		}
		return s
	}
	for _, e := range entitlements {
		of(e.CollectionID).entitled[e.UserID] = e
	}
	for _, h := range holders {
		of(h.CollectionID).holders[h.UserID] = true
	}
	return out, nil
}

func transform(collection *entity.Collection, userID string, permission base.CollectionPermission,
	manages bool, s *standing) *collectiondto.CollectionResp {
	resp := &collectiondto.CollectionResp{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		KeyVersion:  collection.KeyVersion,
		Permission:  permission,
	}
	if s == nil {
		return resp
	}
	resp.HasKey = s.holders[userID]
	if manages {
		resp.Waiting = s.waiting()
		resp.NeedsNewKey = s.needsNewKey()
	}
	return resp
}

// List returns the collections the person may use, and to an admin every one of the organization.
func (uc *UC) List(ctx context.Context, auth *basedto.Auth, orgID string) (
	*collectiondto.ListCollectionsResp, error) {
	member, err := uc.memberOf(ctx, uc.db, auth, orgID)
	if err != nil {
		return nil, err
	}
	collections, err := uc.collectionRepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	standings, err := uc.standings(ctx, uc.db, orgID, "")
	if err != nil {
		return nil, err
	}
	data := make([]*collectiondto.CollectionResp, 0, len(collections))
	for _, collection := range collections {
		s := standings[collection.ID]
		var permission base.CollectionPermission
		if s != nil && s.entitled[auth.UserID()] != nil {
			permission = s.entitled[auth.UserID()].Permission
		}
		if permission == "" && !isAdmin(member) {
			continue
		}
		manages := isAdmin(member) || permission == base.CollectionPermissionManage
		data = append(data, transform(collection, auth.UserID(), permission, manages, s))
	}
	return &collectiondto.ListCollectionsResp{Data: data}, nil
}

// Get returns a collection, and to those who manage its access its grants and who has access, with their identities
// and whether they hold its key.
func (uc *UC) Get(ctx context.Context, auth *basedto.Auth, orgID, collectionID string) (
	*collectiondto.GetCollectionResp, error) {
	c, err := uc.reach(ctx, uc.db, auth, orgID, collectionID, false)
	if err != nil {
		return nil, err
	}
	standings, err := uc.standings(ctx, uc.db, orgID, collectionID)
	if err != nil {
		return nil, err
	}
	s := standings[collectionID]
	detail := &collectiondto.CollectionDetailResp{
		Collection: transform(c.collection, auth.UserID(), c.permission, c.managesAccess(), s),
		Grants:     []*collectiondto.GrantResp{},
		Members:    []*collectiondto.MemberResp{},
	}
	if !c.managesAccess() {
		return &collectiondto.GetCollectionResp{Data: detail}, nil
	}
	grantees, err := uc.collectionRepo.ListGrantees(ctx, uc.db, orgID, collectionID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, g := range grantees {
		detail.Grants = append(detail.Grants, collectiondto.TransformGrant(g))
	}
	if detail.Members, err = uc.members(ctx, uc.db, orgID, collectionID, s); err != nil {
		return nil, err
	}
	return &collectiondto.GetCollectionResp{Data: detail}, nil
}

func (uc *UC) members(ctx context.Context, db database.IDB, orgID, collectionID string, s *standing) (
	[]*collectiondto.MemberResp, error) {
	entitlements, err := uc.collectionRepo.Entitlements(ctx, db, orgID, collectionID, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	members := make([]*collectiondto.MemberResp, 0, len(entitlements))
	for _, e := range entitlements {
		identity, err := uc.identityRepo.Get(ctx, db, e.UserID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		members = append(members, transformMember(e, s != nil && s.holders[e.UserID], identity))
	}
	return members, nil
}

func transformMember(e *entity.CollectionEntitlement, hasKey bool,
	identity *entity.UserIdentity) *collectiondto.MemberResp {
	return &collectiondto.MemberResp{
		UserID:     e.UserID,
		Username:   e.Username,
		FullName:   e.FullName,
		Role:       e.Role,
		Permission: e.Permission,
		Served:     e.Served(),
		HasKey:     hasKey,
		Identity:   identitydto.TransformIdentity(identity),
	}
}
