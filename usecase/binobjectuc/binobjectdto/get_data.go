package binobjectdto

import (
	"io"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type GetBinObjectDataReq struct {
	ID   string             `json:"-"`
	Type base.BinObjectType `json:"-"`
}

func NewGetBinObjectDataReq() *GetBinObjectDataReq {
	return &GetBinObjectDataReq{}
}

func (req *GetBinObjectDataReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetBinObjectDataResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *BinObjectDataResp `json:"data"`
}

type BinObjectDataResp struct {
	ContentLength int64
	ContentType   string
	Content       io.ReadCloser
	ExtraHeaders  map[string]string
}
