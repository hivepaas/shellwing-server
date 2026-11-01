package devhelperhandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/devhelperuc"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	devHelperUC *devhelperuc.UC
	sessionUC   *sessionuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	devHelperUC *devhelperuc.UC,
	sessionUC *sessionuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		devHelperUC: devHelperUC,
		sessionUC:   sessionUC,
	}
}
