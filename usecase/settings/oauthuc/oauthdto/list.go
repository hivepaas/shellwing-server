package oauthdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

type ListOAuthReq struct {
	settings.ListSettingReq
	Kind []base.OAuthKind `json:"-" mapstructure:"kind"`
}

func NewListOAuthReq() *ListOAuthReq {
	return &ListOAuthReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListOAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	validators = append(validators, basedto.ValidateSlice(req.Kind, true, 0, base.AllOAuthKinds, "kind")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListOAuthResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*OAuthResp      `json:"data"`
}

func TransformOAuths(
	settings []*entity.Setting,
	input *OAuthTransformInput,
) ([]*OAuthResp, error) {
	resp := make([]*OAuthResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformOAuth(setting, input)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
