package systemeventbusserviceimpl

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
)

func New(
	client rediscache.Client,
	logger logging.Logger,
) systemeventbusservice.Service {
	return &service{
		client:    client,
		logger:    logger,
		listeners: make(map[base.SystemEventType][]chan *entity.SystemEvent),
	}
}

type service struct {
	client rediscache.Client
	logger logging.Logger

	mu        sync.RWMutex
	listeners map[base.SystemEventType][]chan *entity.SystemEvent

	pubSub     *redis.PubSub
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}
