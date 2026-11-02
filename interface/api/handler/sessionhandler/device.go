package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

// AuthorizeDevice lets an app on the person's device sign in
// @Summary Lets an app on the person's device sign in
// @Description The person approves in their browser. The answer is where the browser goes next: the app's loopback
// @Description address, with a code.
// @Tags    Sessions
// @Produce json
// @Id      authorizeDevice
// @Param   body body sessiondto.AuthorizeDeviceReq true "request data"
// @Success 200 {object} sessiondto.AuthorizeDeviceResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /devices/authorize [post]
func (h *Handler) AuthorizeDevice(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := sessiondto.NewAuthorizeDeviceReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.AuthorizeDevice(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ExchangeDeviceCode gives an app its tokens for a code
// @Summary Gives an app its tokens for a code
// @Description The app proves it asked for the code with the PKCE verifier. A code works once.
// @Tags    Sessions
// @Produce json
// @Id      exchangeDeviceCode
// @Param   body body sessiondto.ExchangeDeviceCodeReq true "request data"
// @Success 200 {object} sessiondto.DeviceTokensResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /devices/token [post]
func (h *Handler) ExchangeDeviceCode(ctx *gin.Context) {
	req := sessiondto.NewExchangeDeviceCodeReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.ExchangeDeviceCode(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// RefreshDevice gives an app a new access token
// @Summary Gives an app a new access token
// @Description For the device's refresh token, which stays as it is.
// @Tags    Sessions
// @Produce json
// @Id      refreshDevice
// @Param   body body sessiondto.DeviceRefreshTokenReq true "request data"
// @Success 200 {object} sessiondto.DeviceTokensResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 401 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /devices/refresh [post]
func (h *Handler) RefreshDevice(ctx *gin.Context) {
	req := sessiondto.NewDeviceRefreshTokenReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.RefreshDevice(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// SignOutDevice signs an app out
// @Summary Signs an app out
// @Description With its own refresh token, which stops working.
// @Tags    Sessions
// @Produce json
// @Id      signOutDevice
// @Param   body body sessiondto.DeviceRefreshTokenReq true "request data"
// @Success 200 {object} sessiondto.SignOutDeviceResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /devices/sign-out [post]
func (h *Handler) SignOutDevice(ctx *gin.Context) {
	req := sessiondto.NewDeviceRefreshTokenReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.SignOutDevice(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ListDevices lists the person's devices
// @Summary Lists the person's devices
// @Description The apps signed in as the person.
// @Tags    Sessions
// @Produce json
// @Id      listDevices
// @Success 200 {object} sessiondto.ListDevicesResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /devices [get]
func (h *Handler) ListDevices(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.ListDevices(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// RevokeDevice signs out one of the person's devices
// @Summary Signs out one of the person's devices
// @Description Its tokens stop working at once. Its copy of the data stays on the device.
// @Tags    Sessions
// @Produce json
// @Id      revokeDevice
// @Param   deviceID path string true "device ID"
// @Success 200 {object} sessiondto.RevokeDeviceResp
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /devices/{deviceID} [delete]
func (h *Handler) RevokeDevice(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.sessionUC.RevokeDevice(h.RequestCtx(ctx), auth, ctx.Param("deviceID"))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
