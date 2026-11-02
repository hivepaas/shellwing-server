package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/usecase/system/licenseuc/licensedto"
)

func licenseAccess(action base.ActionType) *permission.GeneralResourceAccessCheck {
	return &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeLicense,
	}
}

// GetLicense Gets the license in force
// @Summary Gets the license in force
// @Description The license in force, this server's ID, and how many of its users' seats are taken.
// @Tags    System
// @Produce json
// @Id      getLicense
// @Success 200 {object} licensedto.GetLicenseResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/license [get]
func (h *Handler) GetLicense(ctx *gin.Context) {
	_, err := h.authHandler.GetCurrentAuth(ctx, licenseAccess(base.ActionTypeRead))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.licenseUC.GetLicense(h.RequestCtx(ctx))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// InstallLicense Installs a license key
// @Summary Installs a license key
// @Description Verifies the key, offline, and puts it in force. One for another server, or past its grace, is refused.
// @Tags    System
// @Produce json
// @Id      installLicense
// @Param   body body licensedto.InstallLicenseReq true "request data"
// @Success 200 {object} licensedto.InstallLicenseResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/license [put]
func (h *Handler) InstallLicense(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, licenseAccess(base.ActionTypeWrite))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := licensedto.NewInstallLicenseReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.licenseUC.InstallLicense(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// RemoveLicense Removes the license key
// @Summary Removes the license key
// @Description The free limit is back. Nobody is removed or locked out: a limit only stops creating.
// @Tags    System
// @Produce json
// @Id      removeLicense
// @Success 200 {object} licensedto.RemoveLicenseResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/license [delete]
func (h *Handler) RemoveLicense(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, licenseAccess(base.ActionTypeDelete))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.licenseUC.RemoveLicense(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
