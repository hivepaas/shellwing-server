package devhelperuc

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) LongRequest(
	ctx context.Context,
	req *devhelperdto.LongRequestReq,
) (*devhelperdto.LongRequestResp, error) {
	time.Sleep(req.Duration.ToDuration())
	return &devhelperdto.LongRequestResp{}, nil
}
