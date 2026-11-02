package entity

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
)

// SSHCA is an organization's SSH certificate authority. Its private key is sealed with the server's data key, as any
// stored secret, and has no JSON name: it never leaves in a response.
type SSHCA struct {
	bun.BaseModel `bun:"table:ssh_cas,alias:ssh_ca"`

	ID         string           `bun:",pk" json:"id"`
	OrgID      string           `json:"orgId"`
	Status     base.SSHCAStatus `json:"status"`
	PublicKey  string           `json:"publicKey"`
	PrivateKey string           `bun:",nullzero" json:"-"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	RetiredAt time.Time `bun:",nullzero" json:"retiredAt,omitzero"`
}

// GetID implements IDEntity interface
func (c *SSHCA) GetID() string {
	return c.ID
}

// SetPrivateKey seals the key for storing.
func (c *SSHCA) SetPrivateKey(plain string) error {
	var field EncryptedField
	field.Set(plain)
	sealed, err := field.GetEncrypted()
	if err != nil {
		return hperrors.Wrap(err)
	}
	c.PrivateKey = sealed
	return nil
}

// GetPrivateKey opens the stored key.
func (c *SSHCA) GetPrivateKey() (string, error) {
	var field EncryptedField
	field.Set(c.PrivateKey)
	plain, err := field.GetPlain()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return plain, nil
}
