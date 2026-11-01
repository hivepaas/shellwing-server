package userhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

// GetUserInviteInfo Gets user invite info
// @Summary Gets user invite info
// @Description Gets user invite info
// @Tags    Users
// @Produce json
// @Id      getUserInviteInfo
// @Success 200 {object} userdto.GetUserInviteInfoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /users/invite/info [get]
func (h *Handler) GetUserInviteInfo(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, "", base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewGetUserInviteInfoReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.GetUserInviteInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// InviteUser Invites a user to system
// @Summary Invites a user to system
// @Description Invites a user to system
// @Tags    Users
// @Produce json
// @Id      inviteUser
// @Param   body body userdto.InviteUserReq true "request data"
// @Success 200 {object} userdto.InviteUserResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /users/invite [post]
func (h *Handler) InviteUser(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, "", base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewInviteUserReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.InviteUser(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RequestUserSignup Requests an account with an email address
// @Summary Requests an account with an email address
// @Description Mails a link that finishes the account through signup-begin and signup-complete, as an
// @Description invitation's does. The answer is the same whether or not the address has an account, so
// @Description nothing is mailed to one in use. `next` is a path on the web app to open once the account is
// @Description finished. Refused with 403 when self sign-up is off, and 429 past 5 requests an hour from
// @Description one address or 3 for one email.
// @Tags    Users
// @Produce json
// @Id      requestUserSignup
// @Param   body body userdto.RequestUserSignupReq true "request data"
// @Success 200 {object} userdto.RequestUserSignupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 429 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /users/signup-request [post]
func (h *Handler) RequestUserSignup(ctx *gin.Context) {
	req := userdto.NewRequestUserSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.RequestUserSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BeginUserSignup Begins user signup process
// @Summary Begins user signup process
// @Description Begins user signup process
// @Tags    Users
// @Produce json
// @Id      beginUserSignup
// @Param   body body userdto.BeginUserSignupReq true "request data"
// @Success 200 {object} userdto.BeginUserSignupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /users/signup-begin [post]
func (h *Handler) BeginUserSignup(ctx *gin.Context) {
	req := userdto.NewBeginUserSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.BeginUserSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CompleteUserSignup Completes user signup process
// @Summary Completes user signup process
// @Description Completes user signup process
// @Tags    Users
// @Produce json
// @Id      completeUserSignup
// @Param   body body userdto.CompleteUserSignupReq true "request data"
// @Success 200 {object} userdto.CompleteUserSignupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /users/signup-complete [post]
func (h *Handler) CompleteUserSignup(ctx *gin.Context) {
	req := userdto.NewCompleteUserSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.CompleteUserSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
