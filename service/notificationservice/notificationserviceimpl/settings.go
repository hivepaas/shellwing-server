package notificationserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

func (s *service) GetNotificationForEvent(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	eventSetting *entity.BaseEventNotification,
	eventSuccess bool,
	refObjects *entity.RefObjects,
) (_ *entity.Notification, err error) {
	if eventSetting == nil {
		return nil, nil
	}
	if refObjects == nil {
		refObjects = entity.NewRefObjects()
	}

	notifSettingID := gofn.If(eventSuccess, eventSetting.Success.ID, eventSetting.Failure.ID)
	if notifSettingID == "" {
		if (eventSuccess && !eventSetting.SuccessUseDefault) || (!eventSuccess && !eventSetting.FailureUseDefault) {
			return nil, nil
		}
	}

	var setting *entity.Setting
	if notifSettingID == "" {
		setting, err = s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
			bunex.SelectWhere("setting.is_default = TRUE"),
		)
	} else {
		if loadedSetting, ok := refObjects.RefSettings[notifSettingID]; ok && loadedSetting != nil {
			setting = loadedSetting
		} else {
			setting, err = s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeNotification,
				notifSettingID, true)
		}
	}
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}

	// Load ref objects of the setting (otherwise we will have error of missing ref objects)
	err = s.settingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, scope, true, setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refObjects.RefSettings[setting.ID] = setting

	return setting.MustAsNotification(), nil
}

func (s *service) GetDefaultNotification(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	refObjects *entity.RefObjects,
	skipMissing bool,
) (*entity.Notification, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
		bunex.SelectWhere("setting.is_default = TRUE"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}

	if refObjects != nil {
		// Load ref objects of the setting (otherwise we will have error of missing ref objects)
		loadFunc := gofn.If(skipMissing, s.settingService.LoadRefObjectsSkipMissing, s.settingService.LoadRefObjects)
		err = loadFunc(ctx, db, &refObjects, scope, true, setting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		refObjects.RefSettings[setting.ID] = setting
	}

	return setting.MustAsNotification(), nil
}

func (s *service) BuildTitlePrefix(
	user *entity.User,
) string {
	switch {
	case user != nil:
		return fmt.Sprintf("[User][%s]", gofn.Coalesce(user.FullName, user.Username))
	default:
		return "[System]"
	}
}

func (s *service) BuildTitlePrefixForScope(
	scope *entity.ObjectScope,
) string {
	switch scope.ScopeType {
	case base.ObjectScopeUser:
		return fmt.Sprintf("[User][%s]", gofn.Coalesce(scope.User.FullName, scope.User.Username))
	case base.ObjectScopeGlobal:
		return "[System]"
	}
	return ""
}
