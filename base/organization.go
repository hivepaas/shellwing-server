package base

// DefaultOrgSlug names the organization a self-hosted server has.
const DefaultOrgSlug = "default"

type OrgStatus string

const (
	OrgStatusActive   OrgStatus = "active"
	OrgStatusDisabled OrgStatus = "disabled"
)

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

// OrgMemberStatus is what the organization says of a member. Whether their account may sign in is the user's status.
type OrgMemberStatus string

const (
	OrgMemberStatusActive OrgMemberStatus = "active"
	// OrgMemberStatusSuspended is a member an admin stopped for a while: the organization gives them nothing, and
	// keeps what they were granted.
	OrgMemberStatusSuspended OrgMemberStatus = "suspended"
	// OrgMemberStatusRemoved is a member no more: what they were granted is gone, and they come back as a member.
	OrgMemberStatusRemoved OrgMemberStatus = "removed"
)

var AllOrgRoles = []OrgRole{OrgRoleOwner, OrgRoleAdmin, OrgRoleMember}

var AllOrgMemberStatuses = []OrgMemberStatus{
	OrgMemberStatusActive, OrgMemberStatusSuspended, OrgMemberStatusRemoved,
}

type SpaceType string

const (
	SpaceTypePersonal   SpaceType = "personal"
	SpaceTypeCollection SpaceType = "collection"
)

// CollectionPermission is what a grant lets a person do in a collection. Each takes in the ones before it.
type CollectionPermission string

const (
	// CollectionPermissionConnect sees the hosts and connects.
	CollectionPermissionConnect CollectionPermission = "connect"
	// CollectionPermissionEdit adds and changes records.
	CollectionPermissionEdit CollectionPermission = "edit"
	// CollectionPermissionManage grants others, and replaces the key.
	CollectionPermissionManage CollectionPermission = "manage"
)

var AllCollectionPermissions = []CollectionPermission{
	CollectionPermissionConnect, CollectionPermissionEdit, CollectionPermissionManage,
}

// Rank orders permissions: 0 for none, then connect, edit, manage.
func (p CollectionPermission) Rank() int {
	switch p {
	case CollectionPermissionConnect:
		return 1
	case CollectionPermissionEdit:
		return 2 //nolint:mnd
	case CollectionPermissionManage:
		return 3 //nolint:mnd
	}
	return 0
}

// AtLeast says whether the permission takes in the other.
func (p CollectionPermission) AtLeast(other CollectionPermission) bool {
	return p.Rank() >= other.Rank() && p.Rank() > 0
}

type SSHCAStatus string

const (
	SSHCAStatusActive  SSHCAStatus = "active"
	SSHCAStatusRetired SSHCAStatus = "retired"
)

// The bounds of how long an SSH certificate lasts. Certificates are not revoked, they expire: this keeps that soon.
const (
	SSHCertTTLMinMinutes = 5
	SSHCertTTLMaxMinutes = 1440
)
