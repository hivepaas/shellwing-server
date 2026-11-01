package authhandler

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc"
)

var (
	NoAccessCheck = (permission.AccessCheck)(nil)
)

type Handler struct {
	*handler.BaseHandler
	sessionUC sessionUseCase
}

// sessionUseCase is what the handler asks of sessionuc.UC, narrowed so that a
// test can stand in for it.
type sessionUseCase interface {
	GetCurrentUserByJWT(ctx context.Context, token string) (*basedto.User, error)
	GetCurrentAuthByJWT(ctx context.Context, token string) (*basedto.Auth, error)
	VerifyAuth(ctx context.Context, auth *basedto.Auth, accessCheck permission.AccessCheck) error
}

func New(
	baseHandler *handler.BaseHandler,
	sessionUC *sessionuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		sessionUC:   sessionUC,
	}
}

func (h *Handler) GetCurrentUser(ctx *gin.Context) (*basedto.User, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = refuseDemo(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (h *Handler) currentUser(ctx *gin.Context) (*basedto.User, error) {
	token, err := h.getAuthToken(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if token != "" {
		user, err := h.sessionUC.GetCurrentUserByJWT(h.RequestCtx(ctx), token)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return user, nil
	}

	return nil, hperrors.Wrap(hperrors.ErrNoSession)
}

func (h *Handler) GetCurrentUserByToken(ctx *gin.Context, token string) (*basedto.User, error) {
	user, err := h.sessionUC.GetCurrentUserByJWT(h.RequestCtx(ctx), token)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return user, nil
}

func (h *Handler) GetCurrentAuth(ctx *gin.Context, accessCheck permission.AccessCheck) (*basedto.Auth, error) {
	auth, err := h.getCurrentAuth(ctx)
	if err != nil {
		return auth, hperrors.Wrap(err) // NOTE: on error, still return `auth`
	}

	if err = h.sessionUC.VerifyAuth(ctx, auth, accessCheck); err != nil {
		// NOTE: even on error, we still return the `auth` object so the client code
		// still can be able to check permission with another method.
		return auth, hperrors.Wrap(err)
	}

	return auth, nil
}

// getCurrentAuth is the caller's auth. A request the demo user may not make is
// refused with no auth: nothing is left for a handler to fall back on.
func (h *Handler) getCurrentAuth(ctx *gin.Context) (*basedto.Auth, error) {
	auth, err := h.resolveAuth(ctx)
	if err != nil {
		return auth, err
	}
	if err = refuseDemo(ctx, auth.User); err != nil {
		return nil, err
	}
	return auth, nil
}

func (h *Handler) resolveAuth(ctx *gin.Context) (*basedto.Auth, error) {
	token, err := h.getAuthToken(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if token != "" {
		auth, err := h.sessionUC.GetCurrentAuthByJWT(h.RequestCtx(ctx), token)
		if err != nil {
			return auth, hperrors.Wrap(err) // NOTE: on error, still return `auth`
		}
		return auth, nil
	}

	return nil, hperrors.Wrap(hperrors.ErrNoSession)
}

// getAuthToken gets token from request header `Authorization`.
// The value should be in form of `Bearer <token-data>`.
func (h *Handler) getAuthToken(ctx *gin.Context) (token string, err error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		tokenParts := strings.SplitN(authHeader, " ", 2) //nolint:mnd
		if len(tokenParts) != 2 || tokenParts[1] == "" {
			return "", hperrors.Wrap(hperrors.ErrSessionJWTInvalid)
		}
		return tokenParts[1], nil
	}

	wsProtoHeader := ctx.GetHeader("Sec-WebSocket-Protocol")
	if wsProtoHeader != "" {
		// Header has format: "some_proto, access_token, <token>"
		parts := strings.Split(wsProtoHeader, ",")
		for i := 0; i < len(parts)-1; i++ {
			key := strings.TrimSpace(parts[i])
			if key == "access_token" {
				return strings.TrimSpace(parts[i+1]), nil // The token is the next element
			}
		}
	}

	return "", nil
}
