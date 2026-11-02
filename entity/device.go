package entity

import "time"

// Device is a signed-in app. Its refresh token is known here only by its hash.
type Device struct {
	ID               string `bun:",pk" json:"id"`
	UserID           string `json:"userId"`
	Name             string `json:"name"`
	Platform         string `bun:",nullzero" json:"platform,omitempty"`
	RefreshTokenHash string `bun:",nullzero" json:"-"`

	CreatedAt       time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt       time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	LastSeenAt      time.Time `bun:",nullzero" json:"lastSeenAt,omitzero"`
	RefreshExpireAt time.Time `bun:",nullzero" json:"refreshExpireAt,omitzero"`
	RevokedAt       time.Time `bun:",nullzero" json:"revokedAt,omitzero"`
}

// GetID implements IDEntity interface
func (d *Device) GetID() string {
	return d.ID
}

// CanRefresh says whether the device may still get new tokens at the time given.
func (d *Device) CanRefresh(now time.Time) bool {
	return d.RevokedAt.IsZero() && d.RefreshTokenHash != "" && now.Before(d.RefreshExpireAt)
}
