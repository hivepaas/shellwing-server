package appentity

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
)

type MFATokenClaims struct {
	jwtsession.BaseClaims
	Kind            string       `json:"kind"`
	UserID          string       `json:"userId"`
	MFAType         base.MFAType `json:"mfaType"`
	TrustedDeviceID string       `json:"deviceId,omitempty"`
}
