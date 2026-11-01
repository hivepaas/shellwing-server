package syserroruc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
)

type UC struct {
	db *database.DB

	appErrorRepo repository.SysErrorRepo

	// floodGuard collapses repeats of the same error; see flood_guard.go.
	floodGuard *floodGuard
}

func New(
	db *database.DB,

	appErrorRepo repository.SysErrorRepo,
) *UC {
	return &UC{
		db: db,

		appErrorRepo: appErrorRepo,

		floodGuard: newFloodGuard(),
	}
}
