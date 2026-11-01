package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type UpdateOAuthStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateOAuthStatusReq() *UpdateOAuthStatusReq {
	return &UpdateOAuthStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateOAuthStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingStatusReq.Validate()...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateOAuthStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
