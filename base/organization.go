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
