package devhelperdto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type SimulateCrashReq struct {
}

func NewSimulateCrashReq() *SimulateCrashReq {
	return &SimulateCrashReq{}
}

func (req *SimulateCrashReq) Validate() hperrors.ValidationErrors {
	return nil
}

type SimulateCrashResp struct {
	Meta *basedto.Meta `json:"meta"`
}
