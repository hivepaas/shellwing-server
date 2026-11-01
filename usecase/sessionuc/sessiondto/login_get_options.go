package sessiondto

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type GetLoginOptionsReq struct {
}

func NewGetLoginOptionsReq() *GetLoginOptionsReq {
	return &GetLoginOptionsReq{}
}

func (req *GetLoginOptionsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetLoginOptionsResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data []*LoginOptionResp `json:"data"`
}

type LoginOptionResp struct {
	Type    base.OAuthKind `json:"type"`
	Name    string         `json:"name"`
	Icon    string         `json:"icon"`
	AuthURL string         `json:"authURL"`
}
