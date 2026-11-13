package entity

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
)

type Organization struct {
	ID        string         `bun:",pk" json:"id"`
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	Status    base.OrgStatus `json:"status"`
	UpdateVer int            `json:"updateVer"`
	// SSHCertTTLMinutes is how long a certificate its CA signs lasts.
	SSHCertTTLMinutes int `bun:"ssh_cert_ttl_minutes,nullzero,default:120" json:"sshCertTtlMinutes"`
	// SSHCertExtensions is what a certificate its CA signs is allowed to do, by sshd's extension names. Empty: it
	// logs in and runs a command, with no terminal and nothing reaching past the server.
	SSHCertExtensions []string `bun:"ssh_cert_extensions,array,nullzero,default:'{permit-pty}'" json:"sshCertExtensions"`
	// SSHKRLTokenHash is what a target server fetches the revocation list with, hashed. It has no JSON name: the
	// token is shown once, when it is made, and never again.
	SSHKRLTokenHash      string    `bun:"ssh_krl_token_hash,nullzero" json:"-"`
	SSHKRLTokenCreatedAt time.Time `bun:"ssh_krl_token_created_at,nullzero" json:"sshKrlTokenCreatedAt,omitzero"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (o *Organization) GetID() string {
	return o.ID
}

// OrgMemberUser is a member as the organization lists them: with who they are, and whether their account is active.
type OrgMemberUser struct {
	UserID        string               `bun:"user_id"`
	Role          base.OrgRole         `bun:"role"`
	Status        base.OrgMemberStatus `bun:"status"`
	SSHPrincipals []string             `bun:"ssh_principals,array"`
	Username      string               `bun:"username"`
	Email         string               `bun:"email,nullzero"`
	FullName      string               `bun:"full_name"`
	AccountStatus base.UserStatus      `bun:"account_status"`
}

type OrgMember struct {
	OrgID  string               `bun:",pk" json:"orgId"`
	UserID string               `bun:",pk" json:"userId"`
	Role   base.OrgRole         `json:"role"`
	Status base.OrgMemberStatus `json:"status"`
	// SSHPrincipals are the accounts an SSH certificate of this person may log in as.
	SSHPrincipals []string `bun:"ssh_principals,array,nullzero,default:'{}'" json:"sshPrincipals"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}
