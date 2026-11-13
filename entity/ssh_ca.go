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

	// KeyProvider says where the private half is, as base.SSHCAKeyProviders names them. base.KMSProviderLocal means
	// PrivateKey holds it, sealed. Anything else means a service holds it and this server only asks it to sign.
	KeyProvider string `bun:"key_provider,nullzero,default:'local'" json:"-"`
	// KeyName and KeyVersion name the key in that service. The version is pinned: a key rotated in the service signs
	// under a new public half, and an authority whose public key moved is no longer the one the servers trust.
	KeyName    string `bun:"key_name,nullzero" json:"-"`
	KeyVersion int    `bun:"key_version,nullzero" json:"-"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	RetiredAt time.Time `bun:",nullzero" json:"retiredAt,omitzero"`
}

// GetID implements IDEntity interface
func (c *SSHCA) GetID() string {
	return c.ID
}

// KeyIsLocal reports whether the private half is here, in PrivateKey, rather than in a service.
//
// An empty provider reads as local: it is what every row made before the column existed has.
func (c *SSHCA) KeyIsLocal() bool {
	return c.KeyProvider == "" || c.KeyProvider == base.KMSProviderLocal
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
