package licenseservice

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/licensesig"
)

// State is the license in force on this server. Key holds the customer's email: never log a State.
type State struct {
	Status         base.LicenseStatus
	InstallationID string
	// Plan is the key's while it holds; empty on the free limit.
	Plan string
	// MaxUsers is the limit in force, or base.NoUserLimit.
	MaxUsers int
	// Key is what the key says whenever it verified, also when it is not in force.
	Key         *licensesig.Payload
	GraceEndsAt time.Time
}

// Service answers for the one thing a license changes: how many users a server may have. A limit only stops creating:
// over it, or when a license ends, nobody is removed, disabled or locked out, and every feature keeps working.
type Service interface {
	// EnsureInstallation gives this server its ID the first time. It is safe at every start, on several instances.
	EnsureInstallation(ctx context.Context, db database.IDB) error
	Current(ctx context.Context, db database.IDB) (*State, error)
	// CountUserSeats is how many seats are taken now.
	CountUserSeats(ctx context.Context, db database.IDB) (int, error)
	// CheckUserSeat refuses one more user past the limit. Call it wherever a user comes to hold a seat, with the
	// transaction that writes them: it holds the count until that ends, so two at once cannot take the last seat.
	CheckUserSeat(ctx context.Context, tx database.Tx) error
	// Install verifies a key and keeps it. One that puts nothing in force here is refused: another server's, or one
	// past its grace.
	Install(ctx context.Context, db database.IDB, keyText string) (*State, error)
	// Remove takes the key out. The free limit is back.
	Remove(ctx context.Context, db database.IDB) (*State, error)
}
