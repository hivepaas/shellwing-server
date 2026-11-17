package internal

import (
	"context"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/pkg/auditsink"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

// InitAuditSink gives up the sink's connection when the server stops.
//
// It is invoked rather than only provided so that the sink is built at startup: a syslog address that is nonsense
// should stop the start, not wait for the first recordable action and refuse that instead.
func InitAuditSink(lc fx.Lifecycle, sink auditsink.Sink, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Infof("audit records go to: %s", sink.Name())
			return nil
		},
		OnStop: func(context.Context) error {
			return sink.Close() //nolint:wrapcheck
		},
	})
}
