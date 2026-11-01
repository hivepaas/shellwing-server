package imagehandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/binobjectuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	binObjectUC *binobjectuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	binObjectUC *binobjectuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		binObjectUC: binObjectUC,
	}
}
