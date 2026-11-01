package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type DeleteCloudStorageReq struct {
	settings.DeleteSettingReq
}

func NewDeleteCloudStorageReq() *DeleteCloudStorageReq {
	return &DeleteCloudStorageReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteCloudStorageReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteCloudStorageResp struct {
	Meta *basedto.Meta `json:"meta"`
}
