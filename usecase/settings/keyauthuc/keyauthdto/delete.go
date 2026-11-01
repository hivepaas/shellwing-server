package keyauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type DeleteKeyAuthReq struct {
	settings.DeleteSettingReq
}

func NewDeleteKeyAuthReq() *DeleteKeyAuthReq {
	return &DeleteKeyAuthReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteKeyAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteKeyAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
