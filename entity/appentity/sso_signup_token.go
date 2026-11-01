package appentity

import (
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
)

// SSOSignupTokenClaims are what a sign-in with a provider learned of somebody who has no account yet: an email the
// provider vouches for, and their name and picture. Confirmed on the sign-up page, they make the account.
type SSOSignupTokenClaims struct {
	jwtsession.BaseClaims
	Kind      string `json:"kind"`
	Email     string `json:"email"`
	FullName  string `json:"fullName,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	// Provider is the OAuth setting the person signed in with.
	Provider string `json:"provider"`
}
