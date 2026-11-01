package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

// GetSSOSignupInfo Shows the account a sign-in with a provider offers
// @Summary Shows the account a sign-in with a provider offers
// @Description A sign-in with a provider whose email no user has, where people may make their own accounts, ends
// @Description with a token for the sign-up page (the sso_signup_token cookie): the email and name it would make
// @Description the account with. 403 when people may not make their own accounts; 400 (ERR_TOKEN_INVALID) for a
// @Description token not valid or run out.
// @Tags    Sessions
// @Produce json
// @Id      getSSOSignupInfo
// @Param   body body sessiondto.GetSSOSignupInfoReq true "request data"
// @Success 200 {object} sessiondto.GetSSOSignupInfoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /auth/sso/signup-info [post]
func (h *Handler) GetSSOSignupInfo(ctx *gin.Context) {
	req := sessiondto.NewGetSSOSignupInfoReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.GetSSOSignupInfo(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CompleteSSOSignup Makes the account a sign-in with a provider offered
// @Summary Makes the account a sign-in with a provider offered
// @Description Makes the account the token offers, once its person confirmed it, and signs it in: the answer is its
// @Description session, as a password's sign-in answers it. The customer account that waits for the email, if
// @Description one does, becomes theirs. 412 (ERR_EMAIL_UNAVAILABLE) when the email has a user already.
// @Tags    Sessions
// @Produce json
// @Id      completeSSOSignup
// @Param   body body sessiondto.CompleteSSOSignupReq true "request data"
// @Success 200 {object} sessiondto.CompleteSSOSignupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /auth/sso/signup [post]
func (h *Handler) CompleteSSOSignup(ctx *gin.Context) {
	req := sessiondto.NewCompleteSSOSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.CompleteSSOSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	h.writeSessionDataToCookies(ctx, resp.Data.Session, true)
	ctx.JSON(http.StatusOK, resp)
}
