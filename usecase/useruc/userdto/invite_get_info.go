package userdto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type GetUserInviteInfoReq struct {
}

func NewGetUserInviteInfoReq() *GetUserInviteInfoReq {
	return &GetUserInviteInfoReq{}
}

func (req *GetUserInviteInfoReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetUserInviteInfoResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *UserInviteInfoResp `json:"data"`
}

type UserInviteInfoResp struct {
	CanSendInviteEmails bool `json:"canSendInviteEmails"`
}
