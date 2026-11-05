package orgservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	// EnsureDefaultOrg makes the organization of a self-hosted server if it is not there yet, and makes every user a
	// member of it. It is safe at every start, on several instances at once.
	EnsureDefaultOrg(ctx context.Context, db database.IDB) (*entity.Organization, error)
	// MakeOwner gives the default organization its owner: the first admin.
	MakeOwner(ctx context.Context, db database.IDB, user *entity.User) error
	// SyncUsers makes these users, as saved just now, members of the default organization if they are not yet, and
	// takes deleted ones out. A member keeps their role and status: those are the organization's, and one an admin
	// removed is not made a member again.
	SyncUsers(ctx context.Context, db database.IDB, users []*entity.User) error
	// RemoveUser takes a deleted user out of the default organization.
	RemoveUser(ctx context.Context, db database.IDB, userID string) error
}
