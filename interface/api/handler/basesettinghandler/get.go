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

type GetSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type GetSettingOption func(*GetSettingOptions)

func GetSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) GetSettingOption {
	return func(opts *GetSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) GetSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...GetSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "itemID")
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
		r := keyauthdto.NewGetKeyAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.KeyAuthUC.GetKeyAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewGetCloudStorageReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.GetCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewGetEmailReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.GetEmail(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewGetIMServiceReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.GetIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewGetNotificationReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.GetNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewGetOAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.GetOAuth(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = hperrors.NewNotImplemented()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
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
