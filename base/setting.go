package base

import "github.com/tiendc/gofn"

type SettingType string

const (
	SettingTypeCloudStorage SettingType = "cloud-storage"
	SettingTypeEmail        SettingType = "email"
	SettingTypeIMService    SettingType = "im-service"
	SettingTypeKeyAuth      SettingType = "key-auth"
	SettingTypeNotification SettingType = "notification"
	SettingTypeOAuth        SettingType = "oauth"
	SettingTypeSchedJob     SettingType = "sched-job"
	SettingTypeSecret       SettingType = "secret"
)

type SettingStatus string

const (
	SettingStatusActive   SettingStatus = "active"
	SettingStatusPending  SettingStatus = "pending"
	SettingStatusDisabled SettingStatus = "disabled"
	SettingStatusExpired  SettingStatus = "expired"
	SettingStatusMissing  SettingStatus = "missing" // this is not used in DB
)

var (
	AllSettingStatuses = []SettingStatus{SettingStatusActive, SettingStatusPending, SettingStatusDisabled,
		SettingStatusExpired}
	AllSettingSettableStatuses = gofn.Drop(AllSettingStatuses, SettingStatusExpired)
)
