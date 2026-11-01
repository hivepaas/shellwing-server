package sessiondto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type DeleteSessionReq struct {
	User *basedto.User `json:"-"`
}

func NewDeleteSessionReq() *DeleteSessionReq {
	return &DeleteSessionReq{}
}

func (req *DeleteSessionReq) Validate() hperrors.ValidationErrors {
	return nil
}

type DeleteSessionResp struct {
	Meta *basedto.Meta `json:"meta"`
}
