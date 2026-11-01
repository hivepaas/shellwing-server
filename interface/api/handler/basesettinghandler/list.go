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

type ListSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type ListSettingOption func(*ListSettingOptions)

func ListSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) ListSettingOption {
	return func(opts *ListSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) ListSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...ListSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &ListSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "")
	default:
		err = hperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var paging *basedto.Paging
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeKeyAuth:
		r := keyauthdto.NewListKeyAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.KeyAuthUC.ListKeyAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewListCloudStorageReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.ListCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewListEmailReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.EmailUC.ListEmail(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewListIMServiceReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.ListIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewListNotificationReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.ListNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewListOAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.ListOAuth(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = hperrors.NewNotImplemented()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Every settings list pages: without this, pageOffset, pageLimit and sort
	// would be read by no one, and every list answered whole.
	if pager, ok := req.(interface{ PagingReq() *basedto.Paging }); ok {
		paging = pager.PagingReq()
	}
	if err = h.ParseAndValidateRequest(ctx, req, paging); err != nil {
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
