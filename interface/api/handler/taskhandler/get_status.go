package taskhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

func (h *Handler) GetTaskStatus(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...GetTaskOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetTaskOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserTasks(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalTasks(ctx, base.ResourceTypeTask, base.ActionTypeRead, "itemID")
	default:
		err = hperrors.NewUnsupported("Task scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewGetTaskStatusReq()
	req.Scope, req.ID = scope, itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := h.TaskUC.GetTaskStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if ctx.ContentType() == "text/plain" {
		ctx.String(http.StatusOK, "status=%v", resp.Data.Status)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
