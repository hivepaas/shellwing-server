package auditloghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/auditloguc/auditlogdto"
)

type ListAuditLogOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type ListAuditLogOption func(*ListAuditLogOptions)

func ListAuditLogPreRequestHandler(fn func(auth *basedto.Auth, req any) error) ListAuditLogOption {
	return func(opts *ListAuditLogOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) ListAuditLog(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...ListAuditLogOption,
) {
	var auth *basedto.Auth
	var err error

	options := &ListAuditLogOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserAuditLogs(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalAuditLogs(ctx, base.ResourceTypeAuditLog, base.ActionTypeRead, "")
	default:
		err = hperrors.NewUnsupported("AuditLog scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := auditlogdto.NewListAuditLogReq()
	req.Scope = scope
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := h.AuditLogUC.ListAuditLog(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
