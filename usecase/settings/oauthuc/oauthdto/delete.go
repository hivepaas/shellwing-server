package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type DeleteOAuthReq struct {
	settings.DeleteSettingReq
}

func NewDeleteOAuthReq() *DeleteOAuthReq {
	return &DeleteOAuthReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteOAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteOAuthResp struct {
	Meta *basedto.Meta `json:"meta"`
}
