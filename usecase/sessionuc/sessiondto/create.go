package sessiondto

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

type BaseCreateSessionReq struct {
	User         *entity.User
	AccessAction *base.AccessActions

	// StartedAt is when the login that began this session happened, for a request
	// that is renewing one. Zero starts a new session from now, which is what a
	// login passes - and also what a session older than the deadline claim gets,
	// so an upgrade does not sign everybody out.
	StartedAt time.Time

	// Method is how the caller proved who they are, and it is what the session's
	// audit entry is filed under.
	//
	// Required, and refused when empty rather than defaulted. Every way into
	// shellwing ends at createSession, so that refusal is what stops a login path
	// added later from being recorded as nothing in particular - or, worse, from
	// being the one door with no record at all.
	Method string
}

type BaseCreateSessionResp struct {
	AccessToken     string    `json:"accessToken"`
	AccessTokenExp  time.Time `json:"accessTokenExp"`
	RefreshToken    string    `json:"refreshToken"`
	RefreshTokenExp time.Time `json:"refreshTokenExp"`
}
