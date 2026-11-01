package sessiondto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

// maxSSOSignupTokenLen bounds the token a provider's sign-in handed the sign-up page.
const maxSSOSignupTokenLen = 4000

// SSOSignupTokenReq is the token a sign-in with a provider handed the sign-up page, for somebody with no account.
type SSOSignupTokenReq struct {
	Token string `json:"token"`
}

func (req *SSOSignupTokenReq) ModifyRequest() error {
	req.Token = strings.TrimSpace(req.Token)
	return nil
}

func (req *SSOSignupTokenReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateStr(&req.Token, true, 1,
		maxSSOSignupTokenLen, "token")...))
}

type GetSSOSignupInfoReq struct {
	SSOSignupTokenReq
}

func NewGetSSOSignupInfoReq() *GetSSOSignupInfoReq {
	return &GetSSOSignupInfoReq{}
}

type GetSSOSignupInfoResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *SSOSignupInfoDataResp `json:"data"`
}

// SSOSignupInfoDataResp is the account a provider's sign-in offers to make: for this email, under this name.
type SSOSignupInfoDataResp struct {
	Email    string `json:"email"`
	FullName string `json:"fullName,omitempty"`
}

type CompleteSSOSignupReq struct {
	SSOSignupTokenReq
}

func NewCompleteSSOSignupReq() *CompleteSSOSignupReq {
	return &CompleteSSOSignupReq{}
}

// CompleteSSOSignupResp is the new account's session, as a password's sign-in answers it.
type CompleteSSOSignupResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *CompleteSSOSignupDataResp `json:"data"`
}

type CompleteSSOSignupDataResp struct {
	Session *BaseCreateSessionResp `json:"session"`
}
