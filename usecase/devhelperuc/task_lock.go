package devhelperuc

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) LockTask(
	ctx context.Context,
	req *devhelperdto.LockTaskReq,
) (*devhelperdto.LockTaskResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		_, err := uc.taskRepo.GetByID(ctx, db, nil, "", req.TaskID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		time.Sleep(req.Duration.ToDuration())
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &devhelperdto.LockTaskResp{}, nil
}
