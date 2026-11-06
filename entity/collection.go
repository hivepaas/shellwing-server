package entity

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
)

// Collection is a space an organization shares. Its ID is its space's.
type Collection struct {
	ID          string `bun:",pk" json:"id"`
	OrgID       string `json:"orgId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// KeyVersion is the version of the key its records are sealed with now: 0 until an app makes the first.
	KeyVersion int    `json:"keyVersion"`
	CreatedBy  string `bun:",nullzero" json:"createdBy,omitempty"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (c *Collection) GetID() string {
	return c.ID
}

// CollectionGrant is a permission in a collection, for a person or a group: one of UserID and GroupID is set.
type CollectionGrant struct {
	ID           string                    `bun:",pk" json:"id"`
	CollectionID string                    `json:"collectionId"`
	OrgID        string                    `json:"orgId"`
	UserID       string                    `bun:",nullzero" json:"userId,omitempty"`
	GroupID      string                    `bun:",nullzero" json:"groupId,omitempty"`
	Permission   base.CollectionPermission `json:"permission"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

// CollectionGrantee is a grant with the name of whom it is for.
type CollectionGrantee struct {
	CollectionGrant `bun:",extend"`

	Username  string `bun:"username"`
	FullName  string `bun:"full_name"`
	GroupName string `bun:"group_name"`
}

// CollectionEntitlement is what a person of the organization may do in a collection, through their role, their grant
// and their groups', with what decides whether they are served now.
type CollectionEntitlement struct {
	CollectionID  string                    `bun:"collection_id"`
	UserID        string                    `bun:"user_id"`
	Username      string                    `bun:"username"`
	FullName      string                    `bun:"full_name"`
	Role          base.OrgRole              `bun:"role"`
	Permission    base.CollectionPermission `bun:"permission"`
	MemberStatus  base.OrgMemberStatus      `bun:"member_status"`
	AccountStatus base.UserStatus           `bun:"account_status"`
}

// Served says whether the server gives them the collection now: suspended people and disabled accounts keep their
// permission, and get nothing.
func (e *CollectionEntitlement) Served() bool {
	return e.MemberStatus == base.OrgMemberStatusActive && e.AccountStatus == base.UserStatusActive
}

// CollectionKeyWrap is a collection's key wrapped for one person, at one version.
type CollectionKeyWrap struct {
	CollectionID    string `bun:",pk"`
	UserID          string `bun:",pk"`
	KeyVersion      int    `bun:",pk"`
	RecipientX25519 []byte `bun:"recipient_x25519"`
	EphemeralPublic []byte `bun:"ephemeral_public"`
	SealedKey       []byte `bun:"sealed_key"`
	Commitment      []byte `bun:"commitment"`
	WrappedBy       string `bun:"wrapped_by"`
	WrapperX25519   []byte `bun:"wrapper_x25519"`
	WrapperEd25519  []byte `bun:"wrapper_ed25519"`
	Signature       []byte `bun:"signature"`
	// WrapperUsername is read with the wrap, never stored.
	WrapperUsername string `bun:"wrapper_username,scanonly"`

	CreatedAt time.Time `bun:",default:current_timestamp"`
}

// CollectionKeyHolder is who holds a collection's current key.
type CollectionKeyHolder struct {
	CollectionID string `bun:"collection_id"`
	UserID       string `bun:"user_id"`
}
