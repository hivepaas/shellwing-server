package userserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

func (s *service) LoadNotificationUsers(
	ctx context.Context,
	db database.IDB,
	loadAdmins bool,
) (map[string]*entity.User, error) {
	if !loadAdmins {
		return nil, nil
	}
	userIDs := make([]string, 0, 10) //nolint:mnd

	userMap, err := s.LoadUsersCustomConds(ctx, db, false,
		bunex.SelectWhere("\"user\".id IN (?)", bunex.List(userIDs)),
		bunex.SelectWhereOrIf(loadAdmins, "\"user\".role = ?", base.UserRoleAdmin),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return userMap, nil
}
