package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc/notificationdto"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc/oauthdto"
)

type UpdateSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type UpdateSettingOption func(*UpdateSettingOptions)

func UpdateSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) UpdateSettingOption {
	return func(opts *UpdateSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) UpdateSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...UpdateSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &UpdateSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "itemID")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeKeyAuth:
		r := keyauthdto.NewUpdateKeyAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.KeyAuthUC.UpdateKeyAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewUpdateCloudStorageReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.UpdateCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewUpdateEmailReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.UpdateEmail(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewUpdateIMServiceReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.UpdateIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewUpdateNotificationReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.UpdateNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewUpdateOAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.UpdateOAuth(reqCtx, auth, r) }
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := ucFunc()
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
