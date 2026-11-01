package appentity

import (
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
)

type UserInviteTokenClaims struct {
	jwtsession.BaseClaims
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
}
