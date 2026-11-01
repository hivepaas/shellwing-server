package settingservice

import "github.com/hivepaas/shellwing-server/entity"

type PersistingSettingData struct {
	UpsertingSettings []*entity.Setting
	UpsertingAccesses []*entity.ACLPermission
	DeletingAccesses  []*entity.ACLPermission
}
