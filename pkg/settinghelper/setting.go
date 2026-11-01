package settinghelper

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

func FindSettingsByType(
	settings []*entity.Setting,
	typ base.SettingType,
) (resp []*entity.Setting) {
	for _, setting := range settings {
		if setting.Type == typ {
			resp = append(resp, setting)
		}
	}
	return resp
}

func FindSettingByType(
	settings []*entity.Setting,
	typ base.SettingType,
) *entity.Setting {
	for _, setting := range settings {
		if setting.Type == typ {
			return setting
		}
	}
	return nil
}

func FindSettingsByScope(
	settings []*entity.Setting,
	scope *entity.ObjectScope,
	typ base.SettingType,
	kind *string,
) (res []*entity.Setting) {
	var settingsInParentApp, settingsInProjectEnv, settingsInProject, settingsInGlobal []*entity.Setting
	scopeObjectID := scope.ScopeObjectID()
	for _, setting := range settings {
		if setting.Type != typ || (kind != nil && setting.Kind != *kind) {
			continue
		}
		if setting.ObjectID == "" {
			settingsInGlobal = append(settingsInGlobal, setting)
			continue
		}
		if setting.ObjectID == scopeObjectID {
			res = append(res, setting) // found a setting in current scope
			continue
		}
		switch scope.ScopeType {
		case base.ObjectScopeUser:
		case base.ObjectScopeGlobal:
		}
	}
	res = append(res, settingsInParentApp...)
	res = append(res, settingsInProjectEnv...)
	res = append(res, settingsInProject...)
	res = append(res, settingsInGlobal...)
	return res
}
