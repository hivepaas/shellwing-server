package settinginitserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

func (s *service) initDefaultNotificationSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	notifSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeGlobal,
		Type:        base.SettingTypeNotification,
		Status:      base.SettingStatusActive,
		Name:        "default",
		Inheritable: false,
		Default:     true,
		Version:     entity.CurrentNotificationVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	notifSetting.MustSetData(entity.NewNotificationDefaultForScope(entity.NewObjectScopeGlobal()))

	err = s.settingRepo.Insert(ctx, db, notifSetting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
