package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// SSHCertificate is what one signed certificate says. The certificate itself is not stored: this is the record that
// lets it be named in a revocation list once it should not be used any more.
//
// Serial is signed here, and in the database, because pkg/sshca draws one within 63 bits for exactly that reason.
type SSHCertificate struct {
	bun.BaseModel `bun:"table:ssh_certificates,alias:ssh_cert"`

	Serial      int64    `bun:",pk" json:"serial"`
	OrgID       string   `json:"orgId"`
	CAID        string   `bun:"ca_id" json:"caId"`
	UserID      string   `json:"userId"`
	DeviceID    string   `json:"deviceId"`
	Fingerprint string   `json:"fingerprint"`
	KeyID       string   `bun:"key_id" json:"keyId"`
	Principals  []string `bun:",array" json:"principals"`

	ValidAfter  time.Time `json:"validAfter"`
	ValidBefore time.Time `json:"validBefore"`
	RevokedAt   time.Time `bun:",nullzero" json:"revokedAt,omitzero"`
	CreatedAt   time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// Revoked says whether the certificate has been refused.
func (c *SSHCertificate) Revoked() bool {
	return !c.RevokedAt.IsZero()
}
