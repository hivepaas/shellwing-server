package queueimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/infra/gocronqueue"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/safego"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/settingservice"
	"github.com/hivepaas/shellwing-server/service/startupservice"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
	"github.com/hivepaas/shellwing-server/service/taskservice"
	"github.com/hivepaas/shellwing-server/tasks/queue"
)

type taskQueue struct {
	db          *database.DB
	config      *config.Config
	logger      logging.Logger
	server      *gocronqueue.Server
	client      *gocronqueue.Client
	redisClient rediscache.Client

	settingRepo  repository.SettingRepo
	taskRepo     repository.TaskRepo
	taskInfoRepo cacherepository.TaskInfoRepo

	settingService settingservice.Service
	startupService startupservice.Service
	systemEventBus systemeventbusservice.Service
	taskService    taskservice.Service

	taskExecutorMap map[base.TaskType]gocronqueue.TaskExecFunc
	rawExecutors    map[base.TaskType]queue.TaskExecFunc
}

func New(
	db *database.DB,
	config *config.Config,
	logger logging.Logger,
	redisClient rediscache.Client,

	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	cacheTaskInfoRepo cacherepository.TaskInfoRepo,

	settingService settingservice.Service,
	startupService startupservice.Service,
	systemEventBus systemeventbusservice.Service,
	taskService taskservice.Service,
) queue.TaskQueue {
	return &taskQueue{
		db:          db,
		config:      config,
		logger:      logger,
		redisClient: redisClient,

		settingRepo:  settingRepo,
		taskRepo:     taskRepo,
		taskInfoRepo: cacheTaskInfoRepo,

		settingService: settingService,
		startupService: startupService,
		systemEventBus: systemEventBus,
		taskService:    taskService,
	}
}

func (q *taskQueue) Start() (err error) {
	runWorker := true

	// Initialize task queue worker if configured
	if runWorker {
		q.logger.Infof("starting task queue worker...")

		q.server, err = gocronqueue.NewServer(&gocronqueue.Config{
			TaskMap:             q.taskExecutorMap,
			RedisClient:         q.redisClient,
			Logger:              q.logger,
			Concurrency:         config.Current().Tasks.Queue.Concurrency,
			TaskCheckInterval:   config.Current().Tasks.Queue.TaskCheckInterval,
			TaskCheckFunc:       q.findSchedulingTasks,
			TaskCreateInterval:  config.Current().Tasks.Queue.TaskCreateInterval,
			TaskCreateFunc:      q.doCreateTasksForJobs,
			TaskCanScheduleFunc: q.canScheduleTask,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		// NOTE: use a goroutine-local error, assigning to the outer `err` from
		// here would race with the rest of this function.
		safego.GoWithLogger(q.logger, "taskQueue.serverStart", func() {
			if startErr := q.server.Start(); startErr != nil {
				q.logger.Errorf("failed to start task queue worker: %v", startErr)
			}
		})
	}

	// Initialize task queue client (always init a task queue client)
	q.logger.Infof("starting task queue client...")
	q.client, err = gocronqueue.NewClient(q.redisClient, q.logger)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (q *taskQueue) Shutdown() error {
	q.logger.Info("stopping task queue ...")
	if q.server != nil {
		if err := q.server.Shutdown(); err != nil {
			q.logger.Errorf("failed to stop task queue server: %v", err)
			return hperrors.Wrap(err)
		}
	}
	if q.client != nil {
		if err := q.client.Close(); err != nil {
			q.logger.Errorf("failed to stop task queue client: %v", err)
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (q *taskQueue) StartScheduler() error {
	if q.server == nil {
		q.logger.Error("task queue server is not running")
		return hperrors.Wrap(hperrors.ErrUnavailable).WithParam("Name", "Task queue server")
	}
	if err := q.server.StartScheduler(); err != nil {
		q.logger.Errorf("failed to start scheduler in task queue server: %v", err)
		return hperrors.Wrap(err)
	}
	return nil
}

func (q *taskQueue) StartAllSchedulers() error {
	if q.client != nil {
		if err := q.client.StartScheduler(context.Background()); err != nil {
			q.logger.Errorf("failed to send start scheduler message to servers: %v", err)
			return hperrors.Wrap(err)
		}
	}
	if q.server != nil {
		return q.StartScheduler()
	}
	return nil
}

// StopScheduler pauses this process's scheduler, if it runs one.
func (q *taskQueue) StopScheduler() error {
	if q.server == nil {
		return nil
	}
	if err := q.server.StopScheduler(); err != nil {
		q.logger.Errorf("failed to stop scheduler in task queue server: %v", err)
		return hperrors.Wrap(err)
	}
	return nil
}

func (q *taskQueue) StopAllSchedulers() error {
	if q.client != nil {
		if err := q.client.StopScheduler(context.Background()); err != nil {
			q.logger.Errorf("failed to send stop scheduler message to servers: %v", err)
			return hperrors.Wrap(err)
		}
	}
	if q.server != nil {
		return q.StopScheduler()
	}
	return nil
}

func (q *taskQueue) isWorkerMode() bool {
	return q.config.RunMode == config.RunModeWorker || q.config.RunMode == config.RunModeAppAndWorker
}
