package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

type UpdatePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func NewUpdatePasswordReq() *UpdatePasswordReq {
	return &UpdatePasswordReq{}
}

func (req *UpdatePasswordReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.CurrentPassword, true,
		1, userservice.PasswordRequirements.MaxLen, "currentPassword")...)
	validators = append(validators, basedto.ValidateStr(&req.NewPassword, true,
		userservice.PasswordRequirements.MinLen, userservice.PasswordRequirements.MaxLen, "newPassword")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdatePasswordResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// NULAllowed implements basedto.NULAllowed: a password is only hashed and compared.
func (req *UpdatePasswordReq) NULAllowed() {}
