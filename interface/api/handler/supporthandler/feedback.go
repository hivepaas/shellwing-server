package supporthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/supportuc/supportdto"
)

// CreateFeedback Creates a feedback
// @Summary Creates a feedback
// @Description Creates a feedback
// @Tags    Support
// @Produce json
// @Id      createFeedback
// @Param   body body supportdto.CreateFeedbackReq true "request data"
// @Success 200 {object} supportdto.CreateFeedbackResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /support/feedbacks [post]
func (h *Handler) CreateFeedback(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := supportdto.NewCreateFeedbackReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.supportUC.CreateFeedback(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
