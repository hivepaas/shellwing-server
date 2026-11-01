package auditloghandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/auditloguc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler *authhandler.Handler
	AuditLogUC  *auditloguc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	auditLogUC *auditloguc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		AuthHandler: authHandler,
		AuditLogUC:  auditLogUC,
	}
}
