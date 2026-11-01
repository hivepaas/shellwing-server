package basesettinghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/permission"
)

func (h *Handler) GetAuthGlobalSettings(
	ctx *gin.Context,
	resourceType base.ResourceType,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, itemID string, err error) {
	return h.GetAuthGlobalSettingsAnyAction(ctx, resourceType, []base.ActionType{action}, paramName)
}

func (h *Handler) GetAuthGlobalSettingsAnyAction(
	ctx *gin.Context,
	resourceType base.ResourceType,
	anyActions []base.ActionType,
	paramName string,
) (auth *basedto.Auth, itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}

	accessCheck := &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{},
		Module:          base.ResourceModuleSettings,
		ResourceType:    resourceType,
		ResourceID:      itemID,
	}
	if len(anyActions) == 1 {
		accessCheck.Action = anyActions[0]
	} else {
		accessCheck.AnyOf = anyActions
	}

	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetAuthUserSettings(
	ctx *gin.Context,
	_ base.ActionType,
	paramName string,
) (auth *basedto.Auth, userID, itemID string, err error) {
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		return
	}
	userID = auth.User.ID
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	return
}
