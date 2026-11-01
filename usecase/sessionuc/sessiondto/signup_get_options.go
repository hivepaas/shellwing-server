package sessiondto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

type GetSignupOptionsReq struct {
}

func NewGetSignupOptionsReq() *GetSignupOptionsReq {
	return &GetSignupOptionsReq{}
}

func (req *GetSignupOptionsReq) Validate() hperrors.ValidationErrors {
	return nil
}

// GetSignupOptionsResp is how a person may get an account here, for the sign-in page to offer it or not.
type GetSignupOptionsResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *SignupOptionsDataResp `json:"data"`
}

type SignupOptionsDataResp struct {
	// SelfSignupEnabled says anybody may create an account with an email address they verify
	// (POST /users/signup-request). Off, an account is only made by an admin's invitation.
	SelfSignupEnabled bool `json:"selfSignupEnabled"`
}
