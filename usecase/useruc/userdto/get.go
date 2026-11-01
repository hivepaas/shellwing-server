package userdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type GetUserReq struct {
	ID          string `json:"-"`
	GetAccesses bool   `json:"-" mapstructure:"getAccesses"`
}

func NewGetUserReq() *GetUserReq {
	return &GetUserReq{}
}

func (req *GetUserReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ID, true, "id")...))
}

type GetUserResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *UserDetailsResp `json:"data"`
}

type UserResp struct {
	ID               string                  `json:"id"`
	Username         string                  `json:"username"`
	Email            string                  `json:"email"`
	Role             base.UserRole           `json:"role"`
	Status           base.UserStatus         `json:"status"`
	FullName         string                  `json:"fullName"`
	Photo            string                  `json:"photo"`
	Position         string                  `json:"position"`
	SecurityOption   base.UserSecurityOption `json:"securityOption"`
	MfaTotpActivated bool                    `json:"mfaTotpActivated,omitempty"`
	Notes            string                  `json:"notes,omitempty"`

	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	AccessExpireAt *time.Time `json:"accessExpireAt" copy:",nilonzero"`
	LastAccess     *time.Time `json:"lastAccess" copy:",nilonzero"`
}

type UserDetailsResp struct {
	*UserResp
	ModuleAccesses basedto.ObjectAccessSliceResp `json:"moduleAccesses"`
	Capabilities   basedto.CapabilitySliceResp   `json:"capabilities"`
}

// TransformUserDetails builds the user details response, reporting the user's
// module accesses and capabilities from the loaded ACL rows. Both stay empty when
// the accesses were not loaded.
func TransformUserDetails(user *entity.User) (resp *UserDetailsResp, err error) {
	userResp, err := TransformUser(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp = &UserDetailsResp{
		UserResp: userResp,
	}

	for _, access := range user.Accesses {
		switch access.ResourceType { //nolint:exhaustive
		case base.ResourceTypeModule:
			resp.ModuleAccesses = append(resp.ModuleAccesses, &basedto.ObjectAccessResp{
				NamedObjectResp: basedto.NamedObjectResp{
					ID: access.ResourceID,
				},
				Access: access.Actions,
			})
		case base.ResourceTypeCapability:
			resp.Capabilities = append(resp.Capabilities, base.ResourceCapability(access.ResourceID))
		}
	}

	return resp, nil
}
