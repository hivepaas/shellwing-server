package devhelperuc

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/repository"
)

type UC struct {
	db *database.DB

	taskRepo repository.TaskRepo
}

func New(
	db *database.DB,

	taskRepo repository.TaskRepo,
) *UC {
	return &UC{
		db: db,

		taskRepo: taskRepo,
	}
}
