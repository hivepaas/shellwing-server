package registry

import (
	"fmt"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/pkg/auditsink"
)

// NewAuditSink builds where a copy of each audit record goes, from the configuration.
//
// It does not connect: a collector started a moment after this server is not a reason the server will not start, and
// the first record is what finds out. What stops the start is a sink configured without what it needs to reach
// anything, because that is a mistake no amount of waiting fixes.
func NewAuditSink(cfg *config.Config) (auditsink.Sink, error) {
	if cfg == nil {
		return auditsink.Nop(), nil
	}
	switch cfg.Audit.Sink {
	case "", config.AuditSinkNone:
		return auditsink.Nop(), nil
	case config.AuditSinkSyslog:
		sink, err := auditsink.NewSyslog(cfg.Audit.Syslog())
		if err != nil {
			return nil, fmt.Errorf("audit sink: %w", err)
		}
		return sink, nil
	}
	// config.Audit.validate refuses an unknown name at load, so reaching this means one was added to
	// config.AuditSinks without being added here.
	return nil, fmt.Errorf("audit sink: %w: %q", auditsink.ErrNotConfigured, cfg.Audit.Sink)
}
