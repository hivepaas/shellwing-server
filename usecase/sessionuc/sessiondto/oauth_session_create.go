package sessiondto

import (
	"github.com/markbates/goth"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

// maxTrustedDeviceIDLen bounds the device ID a sign-in with a provider carries through its round trip.
const maxTrustedDeviceIDLen = 200

type CreateOAuthSessionReq struct {
	User *goth.User
	// TrustedDeviceID is the device the person asked to trust when the sign-in began: no passcode is asked on it
	// for a while, as after a password.
	TrustedDeviceID string
}

func NewCreateOAuthSessionReq() *CreateOAuthSessionReq {
	return &CreateOAuthSessionReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateOAuthSessionReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateStr(&req.TrustedDeviceID, false, 1,
		maxTrustedDeviceIDLen, "trustedDeviceId")...))
}

type CreateOAuthSessionResp struct {
	BaseCreateSessionResp
	// MFAToken, when set, is the passcode a user with an authenticator app still owes: no session is made yet, and
	// the passcode is signed in with as after a password (LoginWithPasscode).
	MFAToken string
	// SignupToken, when set, is the account a provider's sign-in offers somebody whose email no user has: no session
	// is made, and the sign-up page asks them to confirm (CompleteSSOSignup).
	SignupToken string
}
