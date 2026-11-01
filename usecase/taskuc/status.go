package taskuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/taskservice"
	"github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

func (uc *UC) GetTaskStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskStatusReq,
) (*taskdto.GetTaskStatusResp, error) {
	getResp, err := uc.taskService.GetTask(ctx, uc.db, &taskservice.GetTaskReq{
		Scope: req.Scope,
		ID:    req.ID,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.GetTaskStatusResp{
		Data: taskdto.TransformTaskStatus(getResp.Task, getResp.TaskInfo),
	}, nil
}
