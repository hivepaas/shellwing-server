package notificationservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	GetNotificationForEvent(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
		eventSetting *entity.BaseEventNotification, eventSuccess bool, refObjects *entity.RefObjects) (
		*entity.Notification, error)
	GetDefaultNotification(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
		refObjects *entity.RefObjects, skipMissing bool) (*entity.Notification, error)

	BuildTitlePrefix(user *entity.User) string
	BuildTitlePrefixForScope(scope *entity.ObjectScope) string

	NotifyForTaskResult(ctx context.Context, db database.IDB, data *TaskResultNotificationReq) (
		*TaskResultNotificationResp, error)
}
