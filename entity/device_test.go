package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	idleWindow = 30 * 24 * time.Hour
	maxAge     = 90 * 24 * time.Hour
)

// signedIn is a device as it is a moment after somebody signed it in.
func signedIn(at time.Time) *Device {
	return &Device{
		ID:               "d1",
		RefreshTokenHash: "a-hash",
		CreatedAt:        at,
		RefreshExpireAt:  at.Add(idleWindow),
	}
}

// The one that matters, and the reason the limit exists. The idle window starts over on every renewal and an app
// renews every few minutes, so no amount of shortening it ever ends a session for a device that is in use - including
// one somebody else is holding. Only the age does.
func TestADeviceInConstantUseStillStopsAtItsAge(t *testing.T) {
	start := time.Now().UTC()
	device := signedIn(start)

	// Renewing every quarter of an hour, as an app with a fifteen minute access token does.
	at := start
	for range 24 * 4 * 80 { // eighty days of it
		at = at.Add(15 * time.Minute)
		if !assert.True(t, device.CanRefresh(at, maxAge), "day %.0f", at.Sub(start).Hours()/24) {
			break
		}
		device.RefreshExpireAt = at.Add(idleWindow)
	}

	assert.True(t, device.CanRefresh(start.Add(maxAge-time.Minute), maxAge), "up to its age")
	assert.False(t, device.CanRefresh(start.Add(maxAge), maxAge), "and not past it")
	assert.False(t, device.CanRefresh(start.Add(maxAge+time.Hour), maxAge))

	// Without the age, the same device goes on for ever, which is what it did before. Renewed a moment ago so that
	// the idle window is wide open, ten years in: the age is the only thing that could stop it, and does.
	far := start.Add(10 * 365 * 24 * time.Hour)
	device.RefreshExpireAt = far.Add(idleWindow)
	assert.False(t, device.CanRefresh(far, maxAge), "ten years of use, with an age limit")
	assert.True(t, device.CanRefresh(far, 0), "and none without one")
}

func TestADeviceLeftUnusedStopsAtItsIdleWindow(t *testing.T) {
	start := time.Now().UTC()
	device := signedIn(start)

	assert.True(t, device.CanRefresh(start.Add(idleWindow-time.Minute), maxAge))
	assert.False(t, device.CanRefresh(start.Add(idleWindow), maxAge), "unused for the whole window")
	assert.False(t, device.CanRefresh(start.Add(idleWindow+time.Hour), 0),
		"and no age limit does not save it: the two bound different things")
}

func TestARevokedDeviceRefreshesNothing(t *testing.T) {
	start := time.Now().UTC()
	device := signedIn(start)
	device.RevokedAt = start.Add(time.Hour)

	assert.False(t, device.CanRefresh(start.Add(2*time.Hour), maxAge))
	assert.False(t, device.CanRefresh(start.Add(2*time.Hour), 0), "revoking does not wait for any window")
}

func TestADeviceWithNoTokenRefreshesNothing(t *testing.T) {
	start := time.Now().UTC()
	device := signedIn(start)
	device.RefreshTokenHash = ""

	assert.False(t, device.CanRefresh(start.Add(time.Hour), maxAge),
		"a device signed out keeps no hash to be found by")
}

// The limits are read at each refresh rather than written into the row, so shortening either one in the configuration
// applies to the devices already signed in.
func TestShorteningTheAgeAppliesToDevicesAlreadySignedIn(t *testing.T) {
	start := time.Now().UTC()
	device := signedIn(start)
	at := start.Add(45 * 24 * time.Hour)
	device.RefreshExpireAt = at.Add(idleWindow)

	assert.True(t, device.CanRefresh(at, maxAge), "ninety days: still fine")
	assert.False(t, device.CanRefresh(at, 30*24*time.Hour), "thirty: over it already")
}
