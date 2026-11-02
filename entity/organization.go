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

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (o *Organization) GetID() string {
	return o.ID
}

type OrgMember struct {
	OrgID  string               `bun:",pk" json:"orgId"`
	UserID string               `bun:",pk" json:"userId"`
	Role   base.OrgRole         `json:"role"`
	Status base.OrgMemberStatus `json:"status"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}
