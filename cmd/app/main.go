package main

import (
	"time"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/cmd/internal"
	"github.com/hivepaas/shellwing-server/registry"
)

const (
	startTimeoutDefault = 30 * time.Second
)

func main() {
	app := fx.New(
		fx.StartTimeout(startTimeoutDefault),
		fx.Provide(registry.Provides...),
		fx.Invoke(internal.InitLogger),
		fx.Invoke(internal.InitConfig),
		fx.Invoke(internal.InitDBConnection),
		fx.Invoke(internal.InitDataKey),
		fx.Invoke(internal.InitCache),
		fx.Invoke(internal.SystemInstallation),
		// After SystemInstallation: the users it creates get the password too.
		fx.Invoke(internal.DevResetUserPasswords),
		fx.Invoke(internal.InitSystemSettings),
		fx.Invoke(internal.InitSystemEventBus),
		fx.Invoke(internal.InitSync),
		fx.Invoke(internal.InitTaskQueue),
		fx.Invoke(internal.InitJWTSession),
		fx.Invoke(internal.InitHTTPServer),
		fx.Invoke(internal.FinalizeStartup),
	)

	app.Run()
}
