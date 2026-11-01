package settinginitserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
)

const (
	lockIDSettingUpdate = "lock:sys:setting-update"
)

func (s *service) InitDefaults(
	ctx context.Context,
	db database.IDB,
) (err error) {
	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		err = s.InitDefaultsWithTx(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) InitDefaultsWithTx(
	ctx context.Context,
	db database.Tx,
) (err error) {
	// Acquire a lock for this process
	_, err = s.lockRepo.GetByID(ctx, db, lockIDSettingUpdate, bunex.SelectFor("UPDATE"))
	if err != nil {
		return hperrors.Wrap(err)
	}

	settings, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectColumns("id", "type", "status"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	// Notification settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeNotification
	}) {
		err = s.initDefaultNotificationSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
