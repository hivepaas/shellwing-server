package devhelperuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) SimulateCrash(
	ctx context.Context,
	req *devhelperdto.SimulateCrashReq,
) (*devhelperdto.SimulateCrashResp, error) {
	panic("app crashes")
}
