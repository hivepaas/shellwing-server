package cloudstorageuc

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeCloudStorage
	currentSettingVersion = entity.CurrentCloudStorageVersion
)

type UC struct {
	*settings.BaseUC
}

func New(
	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		BaseUC: baseUC,
	}
}
