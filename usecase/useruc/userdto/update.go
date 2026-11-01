package userdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/strutil"
)

type UpdateUserReq struct {
	ID             string                       `json:"-"`
	Username       string                       `json:"username"`
	Email          string                       `json:"email"`
	FullName       string                       `json:"fullName"`
	Position       *string                      `json:"position"`
	Status         *base.UserStatus             `json:"status"`
	Role           *base.UserRole               `json:"role"`
	Notes          *string                      `json:"notes"`
	SecurityOption *base.UserSecurityOption     `json:"securityOption"`
	AccessExpireAt *time.Time                   `json:"accessExpireAt"`
	ModuleAccesses basedto.ModuleAccessSliceReq `json:"moduleAccesses"`
	// Capabilities replaces the user's capabilities when present - an empty list
	// takes them all away - and leaves them alone when absent.
	Capabilities basedto.CapabilitySliceReq `json:"capabilities"`
}

func NewUpdateUserReq() *UpdateUserReq {
	return &UpdateUserReq{}
}

func (req *UpdateUserReq) ModifyRequest() error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strutil.NormalizeEmail(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	return nil
}

func (req *UpdateUserReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, validateUsername(&req.Username, false, "username")...)
	validators = append(validators, basedto.ValidateEmail(&req.Email, false, "email")...)
	validators = append(validators, basedto.ValidateStr(&req.FullName, false, nameMinLen, nameMaxLen,
		"fullName")...)
	validators = append(validators, basedto.ValidateStr(req.Position, false,
		nameMinLen, nameMaxLen, "position")...)
	validators = append(validators, basedto.ValidateStrIn(req.Status, false, base.AllUserStatuses,
		"status")...)
	validators = append(validators, basedto.ValidateStrIn(req.Role, false, base.AllUserRoles,
		"role")...)
	validators = append(validators, basedto.ValidateStr(req.Notes, false,
		notesMinLen, notesMaxLen, "notes")...)
	validators = append(validators, basedto.ValidateStrIn(req.SecurityOption, false,
		base.AllUserSecurityOptions, "securityOption")...)
	validators = append(validators, basedto.ValidateModuleAccessSliceReq(req.ModuleAccesses, true,
		0, base.AllResourceModules, "moduleAccesses")...)
	validators = append(validators, basedto.ValidateCapabilitySliceReq(req.Capabilities, true,
		0, base.AllResourceCapabilities, "capabilities")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateUserResp struct {
	Meta *basedto.Meta `json:"meta"`
}
