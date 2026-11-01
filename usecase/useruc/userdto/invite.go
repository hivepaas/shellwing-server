package userdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/strutil"
)

type InviteUserReq struct {
	Email           string                       `json:"email"`
	Role            base.UserRole                `json:"role"`
	SecurityOption  base.UserSecurityOption      `json:"securityOption"`
	AccessExpireAt  time.Time                    `json:"accessExpireAt"`
	SendInviteEmail bool                         `json:"sendInviteEmail"`
	ModuleAccesses  basedto.ModuleAccessSliceReq `json:"moduleAccesses"`
	Capabilities    basedto.CapabilitySliceReq   `json:"capabilities"`
}

func NewInviteUserReq() *InviteUserReq {
	return &InviteUserReq{}
}

func (req *InviteUserReq) ModifyRequest() error {
	req.Email = strutil.NormalizeEmail(req.Email)
	return nil
}

func (req *InviteUserReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStrIn(&req.Role, true,
		base.AllUserRoles, "role")...)
	validators = append(validators, basedto.ValidateStrIn(&req.SecurityOption, true,
		base.AllUserSecurityOptions, "securityOption")...)
	validators = append(validators, validateAdminSecurityOption(req.Role, req.SecurityOption,
		"securityOption")...)
	validators = append(validators, basedto.ValidateModuleAccessSliceReq(req.ModuleAccesses, true,
		0, base.AllResourceModules, "moduleAccesses")...)
	validators = append(validators, basedto.ValidateCapabilitySliceReq(req.Capabilities, true,
		0, base.AllResourceCapabilities, "capabilities")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type InviteUserResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *InviteUserDataResp `json:"data"`
}

type InviteUserDataResp struct {
	InviteLink string `json:"inviteLink"`
}
