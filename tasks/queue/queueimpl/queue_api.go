package queueimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/entityutil"
)

func (q *taskQueue) ScheduleTask(
	ctx context.Context,
	tasks ...*entity.Task,
) error {
	if q.client == nil && q.server == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("task queue is not initialized")
	}

	schedTasks := make([]*entity.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == base.TaskStatusDone || task.Status == base.TaskStatusCanceled {
			continue
		}
		schedTasks = append(schedTasks, task)
	}
	if len(schedTasks) == 0 {
		return nil
	}

	if q.client != nil { // Notify all workers to schedule the tasks
		if err := q.client.ScheduleTask(ctx, schedTasks...); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if q.server != nil { // Notify this worker to schedule the tasks
		if err := q.server.ScheduleTask(ctx, schedTasks...); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (q *taskQueue) UnscheduleTask(
	ctx context.Context,
	tasks ...*entity.Task,
) error {
	if len(tasks) == 0 {
		return nil
	}
	if q.client == nil && q.server == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("task queue is not initialized")
	}

	taskIDs := entityutil.ExtractIDs(tasks)
	if q.client != nil { // Notify all workers to unschedule the tasks
		if err := q.client.UnscheduleTask(ctx, taskIDs...); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if q.server != nil { // Notify this worker to unschedule the tasks
		if err := q.server.UnscheduleTask(ctx, taskIDs...); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
