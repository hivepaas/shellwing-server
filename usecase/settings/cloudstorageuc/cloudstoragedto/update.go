package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type UpdateCloudStorageReq struct {
	settings.UpdateSettingReq
	*CloudStorageBaseReq
}

func NewUpdateCloudStorageReq() *UpdateCloudStorageReq {
	return &UpdateCloudStorageReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateCloudStorageReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateCloudStorageResp struct {
	Meta *basedto.Meta `json:"meta"`
}
