package sessionuc

import (
	"context"
	"errors"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
)

func (uc *UC) GetCurrentUserByJWT(ctx context.Context, jwt string) (*basedto.User, error) {
	authClaims := &jwtsession.AuthClaims{}
	err := jwtsession.ParseToken(jwt, authClaims)
	if err != nil {
		if errors.Is(err, jwtsession.ErrTokenExpired) {
			return nil, hperrors.Wrap(hperrors.ErrSessionJWTExpired).WithCause(err)
		}
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	// Make sure the token is marked `existing` in redis
	if err = uc.userTokenRepo.Exist(ctx, authClaims.UserID, authClaims.UID); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid).WithCause(err)
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, authClaims.UserID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &basedto.User{User: user, AuthClaims: authClaims}, nil
}
