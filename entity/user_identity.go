package entity

import (
	"time"
)

// UserIdentity is a person's identity as their app published it: public keys only.
type UserIdentity struct {
	UserID        string `bun:",pk" json:"userId"`
	X25519Public  []byte `bun:"x25519_public" json:"x25519Public"`
	Ed25519Public []byte `bun:"ed25519_public" json:"ed25519Public"`
	Signature     []byte `json:"signature"`
	Fingerprint   string `json:"fingerprint"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

// GetID implements IDEntity interface
func (i *UserIdentity) GetID() string {
	return i.UserID
}
