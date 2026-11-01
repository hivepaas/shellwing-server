package appentity

import "github.com/hivepaas/shellwing-server/pkg/jwtsession"

type MFATotpSetupTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
	Secret string `json:"secret"`
}
