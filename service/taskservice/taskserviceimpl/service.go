package taskserviceimpl

import (
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/taskservice"
)

func New(
	db *database.DB,
	redisClient rediscache.Client,

	lockRepo repository.LockRepo,
	taskControlRepo cacherepository.TaskControlRepo,
	taskInfoRepo cacherepository.TaskInfoRepo,
	taskLogRepo repository.TaskLogRepo,
	taskRepo repository.TaskRepo,
) taskservice.Service {
	return &service{
		db:          db,
		redisClient: redisClient,

		lockRepo:        lockRepo,
		taskControlRepo: taskControlRepo,
		taskInfoRepo:    taskInfoRepo,
		taskLogRepo:     taskLogRepo,
		taskRepo:        taskRepo,
	}
}

type service struct {
	db          *database.DB
	redisClient rediscache.Client

	lockRepo        repository.LockRepo
	taskControlRepo cacherepository.TaskControlRepo
	taskInfoRepo    cacherepository.TaskInfoRepo
	taskLogRepo     repository.TaskLogRepo
	taskRepo        repository.TaskRepo
}
