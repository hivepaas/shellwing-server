package cacherepository

import (
	"context"
	"fmt"

	"github.com/hivepaas/shellwing-server/entity/cacheentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
	"github.com/hivepaas/shellwing-server/pkg/redishelper"
)

type TaskControlRepo interface {
	Push(ctx context.Context, taskID string, taskControl *cacheentity.TaskControl) error
}

type taskControlRepo struct {
	client rediscache.Client
}

func NewTaskControlRepo(client rediscache.Client) TaskControlRepo {
	return &taskControlRepo{client: client}
}

func (repo *taskControlRepo) Push(
	ctx context.Context,
	taskID string,
	taskControl *cacheentity.TaskControl,
) error {
	key := repo.formatKey(taskID)
	err := redishelper.RPush(ctx, repo.client, key, taskControl)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *taskControlRepo) formatKey(taskID string) string {
	return fmt.Sprintf("task:%s:ctrl", taskID)
}
