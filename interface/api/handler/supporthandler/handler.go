package supporthandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/supportuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	supportUC   *supportuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	supportUC *supportuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		supportUC:   supportUC,
	}
}
