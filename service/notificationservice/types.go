package notificationservice

import (
	"time"

	"github.com/hivepaas/shellwing-server/entity"
)

type TemplateType string

const (
	TemplateTypeEmail   TemplateType = "email"
	TemplateTypeSlack   TemplateType = "slack"
	TemplateTypeDiscord TemplateType = "discord"
)

type TemplateName string

type TemplateData interface {
	GetTitle() string
}

type BaseTemplateData struct {
	Title string
}

func (d *BaseTemplateData) GetTitle() string {
	return d.Title
}

type TaskResultNotificationReq struct {
	ActionSucceeded bool
	Scope           *entity.ObjectScope
	RefObjects      *entity.RefObjects

	Notification *entity.Notification
	TemplateName TemplateName
	TemplateData TemplateData

	LastEvent  string // `success`, `failure`
	LastSendTs time.Time
}

type TaskResultNotificationResp struct {
	SendTs      time.Time
	DeliveryMap map[string]bool
}

func (r *TaskResultNotificationResp) HasSend() bool { // true if has at least one sending
	for _, sent := range r.DeliveryMap {
		if sent {
			return true
		}
	}
	return false
}
