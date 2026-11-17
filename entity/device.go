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
//
// Two limits, and they bound different things. RefreshExpireAt is how long the device may go unused, and it starts
// over on every refresh - so on its own it never arrives for a device that is in use, which is most of them and all
// of the ones somebody else is holding. An app renews every few minutes, so "ninety days" there means ninety days of
// sitting in a drawer, not ninety days of being signed in.
//
// maxAge is measured from when the person signed this device in and does not move. It is what bounds a device nobody
// has reported missing: after it, somebody signs in again however busy the device has been. The other ways a device
// stops working all need somebody to act - revoking it, or the account being disabled - and this one does not.
//
// A maxAge of zero or less is no cap at all, for an installation that wants a device signed in until it is revoked.
func (d *Device) CanRefresh(now time.Time, maxAge time.Duration) bool {
	if !d.RevokedAt.IsZero() || d.RefreshTokenHash == "" || !now.Before(d.RefreshExpireAt) {
		return false
	}
	return maxAge <= 0 || now.Before(d.CreatedAt.Add(maxAge))
}
