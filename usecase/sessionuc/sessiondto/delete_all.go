package sessiondto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type DeleteAllSessionsReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteAllSessionsReq() *DeleteAllSessionsReq {
	return &DeleteAllSessionsReq{}
}

func (req *DeleteAllSessionsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type DeleteAllSessionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
