package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// LicenseState is this server's identity to the portal, and the key installed on it. KeyText holds the customer's
// email: it is never logged.
type LicenseState struct {
	bun.BaseModel `bun:"table:license_state,alias:license_state"`

	ID             int16     `bun:",pk"`
	InstallationID string    ``
	KeyText        string    `bun:",nullzero"`
	UpdatedAt      time.Time `bun:",default:current_timestamp"`
}
