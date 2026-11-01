package devhelperdto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

type LockTaskReq struct {
	TaskID   string            `json:"taskId"`
	Duration timeutil.Duration `json:"duration"`
}

func NewLockTaskReq() *LockTaskReq {
	return &LockTaskReq{}
}

func (req *LockTaskReq) Validate() hperrors.ValidationErrors {
	return nil
}

type LockTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
}
