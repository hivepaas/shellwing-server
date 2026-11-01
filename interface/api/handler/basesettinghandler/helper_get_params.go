package basesettinghandler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetParamGlobalSettings(
	ctx *gin.Context,
	paramName string,
) (itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	return
}
