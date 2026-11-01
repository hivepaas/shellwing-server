package taskuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/taskservice"
	"github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

func (uc *UC) GetTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskReq,
) (*taskdto.GetTaskResp, error) {
	// Within the scope the task was reached through, as its status, logs and
	// cancel are: through one app's path, another app's task is not found.
	getResp, err := uc.taskService.GetTask(ctx, uc.db, &taskservice.GetTaskReq{
		Scope: req.Scope,
		ID:    req.ID,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	refObjects.AddObjectScope(req.Scope)
	err = uc.loadTaskRefData(ctx, uc.db, []*entity.Task{getResp.Task}, &refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTask(getResp.Task, getResp.TaskInfo, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.GetTaskResp{
		Data: resp,
	}, nil
}
