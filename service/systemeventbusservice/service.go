package systemeventbusservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

type Service interface {
	Publish(ctx context.Context, eventType base.SystemEventType, payload ...string) error
	Subscribe(eventTypes ...base.SystemEventType) (<-chan *entity.SystemEvent, func())

	Start(ctx context.Context) error
	Stop() error
}
