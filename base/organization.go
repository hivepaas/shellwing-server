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

type OrgMemberStatus string

const (
	OrgMemberStatusActive   OrgMemberStatus = "active"
	OrgMemberStatusInvited  OrgMemberStatus = "invited"
	OrgMemberStatusDisabled OrgMemberStatus = "disabled"
)

type SpaceType string

const (
	SpaceTypePersonal   SpaceType = "personal"
	SpaceTypeCollection SpaceType = "collection"
)
