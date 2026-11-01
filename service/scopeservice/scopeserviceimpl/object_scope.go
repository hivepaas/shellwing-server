package scopeserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

func (s *service) LoadObjectScopeData(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
) (err error) {
	requireActive := !scope.NotRequireActive
	lock := scope.LockScopeObject
	switch scope.ScopeType {
	case base.ObjectScopeUser:
		scope.User, err = s.userService.LoadUser(ctx, db, scope.UserID, requireActive,
			bunex.SelectForIf(lock, "KEY SHARE OF \"user\""),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

	case base.ObjectScopeGlobal:
		return nil
	}
	return nil
}

func (s *service) LoadObjectScope(
	ctx context.Context,
	db database.IDB,
	scopeType base.ObjectScopeType,
	objectID string,
	requireActive bool,
) (*entity.ObjectScope, error) {
	var scope *entity.ObjectScope
	switch scopeType {
	case base.ObjectScopeUser:
		scope = entity.NewObjectScopeUser(objectID)
	case base.ObjectScopeGlobal:
		scope = entity.NewObjectScopeGlobal()
	}
	// An unrecognized scope type leaves scope nil, and every line below would panic on it. That is
	// reachable whenever the type comes from stored data rather than a validated request.
	if scope == nil {
		return nil, hperrors.NewUnsupported(scopeType)
	}
	scope.NotRequireActive = !requireActive

	err := s.LoadObjectScopeData(ctx, db, scope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return scope, nil
}
