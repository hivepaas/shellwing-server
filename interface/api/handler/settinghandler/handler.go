package settinghandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler/basesettinghandler"
)

type Handler struct {
	*basesettinghandler.Handler
}

func New(
	baseSettingHandler *basesettinghandler.Handler,
) *Handler {
	return &Handler{
		Handler: baseSettingHandler,
	}
}
