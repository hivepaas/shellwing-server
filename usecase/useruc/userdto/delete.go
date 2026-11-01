package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type DeleteUserReq struct {
	ID string `json:"-"`
}

func NewDeleteUserReq() *DeleteUserReq {
	return &DeleteUserReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteUserReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteUserResp struct {
	Meta *basedto.Meta `json:"meta"`
}
