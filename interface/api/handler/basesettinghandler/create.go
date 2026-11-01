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

type CreateSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type CreateSettingOption func(*CreateSettingOptions)

func CreateSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) CreateSettingOption {
	return func(opts *CreateSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) CreateSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...CreateSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &CreateSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
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
		r := keyauthdto.NewCreateKeyAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.KeyAuthUC.CreateKeyAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewCreateCloudStorageReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.CreateCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewCreateEmailReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.EmailUC.CreateEmail(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewCreateIMServiceReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.CreateIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewCreateNotificationReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.CreateNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewCreateOAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.CreateOAuth(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = hperrors.NewNotImplemented()
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

	ctx.JSON(http.StatusCreated, resp)
}
