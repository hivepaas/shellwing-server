package settingserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
)

func (s *service) LoadRefObjects(
	ctx context.Context,
	db database.IDB,
	refObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	inSettings ...*entity.Setting,
) error {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.loadRefObjectsByIDs(ctx, db, refObjects, scope, true, requireActive, allRefIDs)
}

func (s *service) LoadRefObjectsSkipMissing(
	ctx context.Context,
	db database.IDB,
	refObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	inSettings ...*entity.Setting,
) (err error) {
	allRefIDs := &entity.RefObjectIDs{}
	for _, setting := range inSettings {
		allRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	return s.loadRefObjectsByIDs(ctx, db, refObjects, scope, false, requireActive, allRefIDs)
}

func (s *service) LoadRefObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
) error {
	return s.loadRefObjectsByIDs(ctx, db, pRefObjects, scope, true, requireActive, refIDs)
}

func (s *service) LoadRefObjectsByIDsSkipMissing(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
) error {
	return s.loadRefObjectsByIDs(ctx, db, pRefObjects, scope, false, requireActive, refIDs)
}

func (s *service) loadRefObjectsByIDs(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireExistence bool,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
) error {
	requested := &entity.RefObjectIDs{}
	requested.AddRefIDs(refIDs)
	return s.loadRefObjectsByIDsOnce(ctx, db, pRefObjects, scope, requireExistence, requireActive, refIDs, requested)
}

// loadRefObjectsByIDsOnce loads refIDs, then what the loaded settings reference
// that was never asked for: requested holds every ID asked for so far. Leaving
// out only this round's IDs recursed forever on a chain of three settings,
// asking again for the middle one each other round.
func (s *service) loadRefObjectsByIDsOnce(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireExistence bool,
	requireActive bool,
	refIDs *entity.RefObjectIDs,
	requested *entity.RefObjectIDs,
) (err error) {
	if pRefObjects == nil {
		return hperrors.NewArgumentInvalid("refObjects")
	}
	if *pRefObjects == nil {
		*pRefObjects = entity.NewRefObjects()
	}
	refObjects := *pRefObjects

	if refIDs == nil || !refIDs.HasData() {
		return nil
	}

	// Load ref users
	if len(refIDs.RefUserIDs) > 0 {
		err = s.loadReferenceUsers(ctx, db, pRefObjects, requireExistence, requireActive, refIDs.RefUserIDs)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Load ref settings
	if len(refIDs.RefSettingIDs) > 0 {
		err = s.loadReferenceSettings(ctx, db, pRefObjects, scope, requireExistence, requireActive, refIDs.RefSettingIDs)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Calculate recursive ref IDs to load
	newRecursiveRefIDs := requested.GetRecursiveRefObjectIDs(refObjects)
	if !newRecursiveRefIDs.HasData() {
		return nil
	}
	requested.AddRefIDs(newRecursiveRefIDs)

	err = s.loadRefObjectsByIDsOnce(ctx, db, pRefObjects, scope, requireExistence, requireActive,
		newRecursiveRefIDs, requested)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) loadReferenceSettings(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	scope *entity.ObjectScope,
	requireExistence bool,
	requireActive bool,
	settingIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadSettingIDs := make([]string, 0, len(settingIDs))
	for _, settingID := range settingIDs {
		if _, ok := refObjects.RefSettings[settingID]; ok {
			continue
		}
		loadSettingIDs = append(loadSettingIDs, settingID)
	}
	if len(loadSettingIDs) == 0 {
		return nil
	}

	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWithDeleted(),
		bunex.SelectWhereIn("setting.id IN (?)", loadSettingIDs...),
	}
	if requireActive {
		listOpts = append(listOpts, bunex.SelectWhere("setting.status = ?", base.SettingStatusActive))
	}

	settings, _, err := s.settingRepo.List(ctx, db, scope, nil, listOpts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range settings {
		refObjects.RefSettings[setting.ID] = setting
	}

	for _, id := range loadSettingIDs {
		setting := refObjects.RefSettings[id]
		if setting == nil || !setting.DeletedAt.IsZero() {
			if requireExistence {
				return hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("Name", id)
			}
			continue
		}
	}

	return nil
}

func (s *service) loadReferenceUsers(
	ctx context.Context,
	db database.IDB,
	pRefObjects **entity.RefObjects,
	requireExistence bool,
	errorIfUnavailable bool,
	userIDs []string,
) (err error) {
	refObjects := *pRefObjects

	loadUserIDs := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if _, ok := refObjects.RefUsers[userID]; ok {
			continue
		}
		loadUserIDs = append(loadUserIDs, userID)
	}
	if len(loadUserIDs) == 0 {
		return nil
	}

	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWithDeleted(),
	}

	loadUsersFunc := gofn.If(requireExistence, s.userService.LoadUsers, s.userService.LoadUsersSkipMissing)
	userMap, err := loadUsersFunc(ctx, db, loadUserIDs, errorIfUnavailable, listOpts...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, user := range userMap {
		refObjects.RefUsers[user.ID] = user
	}

	return nil
}
