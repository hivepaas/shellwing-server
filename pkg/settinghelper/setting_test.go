package settinghelper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

func TestSetting(t *testing.T) {
	setting1 := &entity.Setting{
		ID:   "s1",
		Type: base.SettingTypeCloudStorage,
		Name: "Cloud Storage 1",
	}
	setting2 := &entity.Setting{
		ID:   "s2",
		Type: base.SettingTypeEmail,
		Name: "Email Setting",
	}
	setting3 := &entity.Setting{
		ID:   "s3",
		Type: base.SettingTypeCloudStorage,
		Name: "Cloud Storage 2",
	}
	settings := []*entity.Setting{setting1, setting2, setting3}

	t.Run("FindSettingsByType", func(t *testing.T) {
		t.Run("multiple matches", func(t *testing.T) {
			res := FindSettingsByType(settings, base.SettingTypeCloudStorage)
			assert.Len(t, res, 2)
			assert.Equal(t, []*entity.Setting{setting1, setting3}, res)
		})

		t.Run("single match", func(t *testing.T) {
			res := FindSettingsByType(settings, base.SettingTypeEmail)
			assert.Len(t, res, 1)
			assert.Equal(t, []*entity.Setting{setting2}, res)
		})

		t.Run("no match", func(t *testing.T) {
			res := FindSettingsByType(settings, base.SettingTypeOAuth)
			assert.Empty(t, res)
		})

		t.Run("empty input", func(t *testing.T) {
			res := FindSettingsByType(nil, base.SettingTypeCloudStorage)
			assert.Empty(t, res)
		})
	})

	t.Run("FindSettingByType", func(t *testing.T) {
		t.Run("match found returns first", func(t *testing.T) {
			res := FindSettingByType(settings, base.SettingTypeCloudStorage)
			assert.NotNil(t, res)
			assert.Equal(t, setting1, res)
		})

		t.Run("single match found", func(t *testing.T) {
			res := FindSettingByType(settings, base.SettingTypeEmail)
			assert.NotNil(t, res)
			assert.Equal(t, setting2, res)
		})

		t.Run("no match returns nil", func(t *testing.T) {
			res := FindSettingByType(settings, base.SettingTypeOAuth)
			assert.Nil(t, res)
		})

		t.Run("empty input returns nil", func(t *testing.T) {
			res := FindSettingByType(nil, base.SettingTypeCloudStorage)
			assert.Nil(t, res)
		})
	})
}
