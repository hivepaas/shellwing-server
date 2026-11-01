package taskuc

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/taskservice"
	"github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

const (
	taskLogBatchThresholdPeriod = time.Millisecond * 1000
	taskLogBatchMaxFrame        = 20
	taskLogSessionTimeout       = 10 * time.Minute
)

func (uc *UC) GetTaskLogs(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.GetTaskLogsReq,
) (*taskdto.GetTaskLogsResp, error) {
	resp, err := uc.taskService.GetTaskLogs(ctx, uc.db, &taskservice.GetTaskLogsReq{
		Scope:                   req.Scope,
		TaskID:                  req.ID,
		Follow:                  req.Follow,
		Since:                   req.Since,
		Duration:                req.Duration.ToDuration(),
		Tail:                    req.Tail,
		LogBatchThresholdPeriod: taskLogBatchThresholdPeriod,
		LogBatchMaxFrame:        taskLogBatchMaxFrame,
		LogSessionTimeout:       taskLogSessionTimeout,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &taskdto.GetTaskLogsResp{
		Data: &taskdto.TaskLogsDataResp{
			StaticLogs:       resp.StaticLogs,
			LogsStream:       resp.LogsStream,
			LogsStreamCloser: resp.LogsStreamCloser,
		},
	}, nil
}
