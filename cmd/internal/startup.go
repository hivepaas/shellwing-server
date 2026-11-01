package internal

import (
	"time"

	"github.com/hivepaas/shellwing-server/pkg/safego"
	"github.com/hivepaas/shellwing-server/service/startupservice"
)

func FinalizeStartup(
	startupService startupservice.Service,
) {
	safego.Go("finalizeStartup", func() {
		time.Sleep(5 * time.Second) //nolint:mnd
		startupService.Shutdown()
	})
}
