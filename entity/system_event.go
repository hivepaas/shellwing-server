package entity

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
)

type SystemEvent struct {
	Type      base.SystemEventType `json:"type"`
	Source    string               `json:"source,omitempty"`
	Payload   string               `json:"payload,omitempty"`
	CreatedAt time.Time            `json:"createdAt"`
}
