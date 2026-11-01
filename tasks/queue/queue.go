package queue

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type TaskQueue interface {
	Start() error
	Shutdown() error

	// The All variants also send to the other processes, but what is sent is
	// read by one process only: whichever gets to it first.
	StartScheduler() error     // resume from pause
	StartAllSchedulers() error // resume from pause
	StopScheduler() error      // pause the scheduler
	StopAllSchedulers() error  // pause the scheduler

	RegisterExecutor(typ base.TaskType, execFunc TaskExecFunc)
	ExecuteTaskType(ctx context.Context, db database.Tx, typ base.TaskType, execData *TaskExecData) error

	ScheduleTask(ctx context.Context, tasks ...*entity.Task) error
	UnscheduleTask(ctx context.Context, tasks ...*entity.Task) error
}
