package binobjectuc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
)

type UC struct {
	db *database.DB

	binObjectRepo repository.BinObjectRepo
}

func New(
	db *database.DB,

	binObjectRepo repository.BinObjectRepo,
) *UC {
	return &UC{
		db: db,

		binObjectRepo: binObjectRepo,
	}
}
