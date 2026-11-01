package taskhandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/taskuc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler *authhandler.Handler
	TaskUC      *taskuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	taskUC *taskuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		AuthHandler: authHandler,
		TaskUC:      taskUC,
	}
}
