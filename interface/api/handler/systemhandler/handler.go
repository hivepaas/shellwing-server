package systemhandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/auditloghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/taskhandler"
	"github.com/hivepaas/shellwing-server/usecase/system/licenseuc"
	"github.com/hivepaas/shellwing-server/usecase/system/syserroruc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler     *authhandler.Handler
	auditLogHandler *auditloghandler.Handler
	taskHandler     *taskhandler.Handler
	sysErrorUC      *syserroruc.UC
	licenseUC       *licenseuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	auditLogHandler *auditloghandler.Handler,
	taskHandler *taskhandler.Handler,
	sysErrorUC *syserroruc.UC,
	licenseUC *licenseuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:     baseHandler,
		authHandler:     authHandler,
		auditLogHandler: auditLogHandler,
		taskHandler:     taskHandler,
		sysErrorUC:      sysErrorUC,
		licenseUC:       licenseUC,
	}
}
