package userhandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/useruc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	userUC      *useruc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	userUC *useruc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		userUC:      userUC,
	}
}
