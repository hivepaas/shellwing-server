package identityhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/identityuc"
	"github.com/hivepaas/shellwing-server/usecase/identityuc/identitydto"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	identityUC  *identityuc.UC
}

func New(baseHandler *handler.BaseHandler, authHandler *authhandler.Handler, identityUC *identityuc.UC) *Handler {
	return &Handler{BaseHandler: baseHandler, authHandler: authHandler, identityUC: identityUC}
}

// GetMine gets the person's identity
// @Summary Gets the person's identity
// @Description The public keys their app published, and the fingerprint people compare. No data when it published
// @Description none yet. For the web app and the person's apps.
// @Tags    Identities
// @Produce json
// @Id      getMyIdentity
// @Success 200 {object} identitydto.GetIdentityResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /users/current/identity [get]
func (h *Handler) GetMine(ctx *gin.Context) {
	auth, err := h.authHandler.GetSyncAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.identityUC.GetMine(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// PutMine publishes the person's identity
// @Summary Publishes the person's identity
// @Description The app publishes the public keys of the identity its vault holds: X25519, Ed25519, and the X25519 key
// @Description signed by the Ed25519 one, base64. A signature that does not check is refused (ERR_IDENTITY_INVALID).
// @Description Another identity takes the place of the one before, and is in the audit log.
// @Tags    Identities
// @Produce json
// @Id      putMyIdentity
// @Param   body body identitydto.PutIdentityReq true "request data"
// @Success 200 {object} identitydto.PutIdentityResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /users/current/identity [put]
func (h *Handler) PutMine(ctx *gin.Context) {
	auth, err := h.authHandler.GetSyncAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := identitydto.NewPutIdentityReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.identityUC.PutMine(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
