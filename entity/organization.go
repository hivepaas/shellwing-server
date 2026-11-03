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
	SSHCertTTLMinutes int `bun:"ssh_cert_ttl_minutes,nullzero,default:480" json:"sshCertTtlMinutes"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (o *Organization) GetID() string {
	return o.ID
}

// OrgMemberUser is a member as an admin lists them: with who they are.
type OrgMemberUser struct {
	UserID        string               `bun:"user_id"`
	Role          base.OrgRole         `bun:"role"`
	Status        base.OrgMemberStatus `bun:"status"`
	SSHPrincipals []string             `bun:"ssh_principals,array"`
	Username      string               `bun:"username"`
	Email         string               `bun:"email,nullzero"`
	FullName      string               `bun:"full_name"`
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
