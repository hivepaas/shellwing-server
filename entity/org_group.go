package entity

import (
	"time"
)

// OrgGroup is people of an organization granted together.
type OrgGroup struct {
	ID          string `bun:",pk" json:"id"`
	OrgID       string `json:"orgId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// SSHPrincipals are the accounts an SSH certificate of each member may log in as.
	SSHPrincipals []string `bun:"ssh_principals,array,nullzero,default:'{}'" json:"sshPrincipals"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

// GetID implements IDEntity interface
func (g *OrgGroup) GetID() string {
	return g.ID
}

type OrgGroupMember struct {
	GroupID string `bun:",pk" json:"groupId"`
	OrgID   string `json:"orgId"`
	UserID  string `bun:",pk" json:"userId"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// OrgGroupMemberUser is a member of a group, with who they are and the group's name.
type OrgGroupMemberUser struct {
	GroupID   string `bun:"group_id"`
	GroupName string `bun:"group_name"`
	UserID    string `bun:"user_id"`
	Username  string `bun:"username"`
	FullName  string `bun:"full_name"`
}
