package emaildto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type DeleteEmailReq struct {
	settings.DeleteSettingReq
}

func NewDeleteEmailReq() *DeleteEmailReq {
	return &DeleteEmailReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteEmailReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteEmailResp struct {
	Meta *basedto.Meta `json:"meta"`
}
