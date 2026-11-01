package devhelperdto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

type LongRequestReq struct {
	Duration timeutil.Duration `json:"duration"`
}

func NewLongRequestReq() *LongRequestReq {
	return &LongRequestReq{}
}

func (req *LongRequestReq) Validate() hperrors.ValidationErrors {
	return nil
}

type LongRequestResp struct {
	Meta *basedto.Meta `json:"meta"`
}
