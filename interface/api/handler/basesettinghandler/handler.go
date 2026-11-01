package basesettinghandler

import (
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler    *authhandler.Handler
	BaseSettingUC  *settings.BaseUC
	KeyAuthUC      *keyauthuc.UC
	CloudStorageUC *cloudstorageuc.UC
	EmailUC        *emailuc.UC
	IMServiceUC    *imserviceuc.UC
	NotificationUC *notificationuc.UC
	OAuthUC        *oauthuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	baseSettingUC *settings.BaseUC,
	keyAuthUC *keyauthuc.UC,
	cloudStorageUC *cloudstorageuc.UC,
	emailUC *emailuc.UC,
	imServiceUC *imserviceuc.UC,
	notificationUC *notificationuc.UC,
	oauthUC *oauthuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:    baseHandler,
		AuthHandler:    authHandler,
		BaseSettingUC:  baseSettingUC,
		KeyAuthUC:      keyAuthUC,
		CloudStorageUC: cloudStorageUC,
		EmailUC:        emailUC,
		IMServiceUC:    imServiceUC,
		NotificationUC: notificationUC,
		OAuthUC:        oauthUC,
	}
}
