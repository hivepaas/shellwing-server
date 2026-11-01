package internal

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/settinginitservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
)

func SystemInstallation(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	userService userservice.Service,
	settingInitService settinginitservice.Service,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sysStatus, err := sysStatusRepo.Get(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to load system status: %w", err)
			}
			config.SetInstallationStep(sysStatus.NextStep)

			if sysStatus.NextStep == base.InstallationStepInitData {
				err = sysInstallationInitData(ctx, db, sysStatusRepo, userService,
					settingInitService, logger)
				if err != nil {
					return fmt.Errorf("failed to initialize system data: %w", err)
				}
			} else {
				// An upgrade gets the defaults its version added - a setting
				// and its job - without waiting for someone to open one. A
				// failure is logged: opening the setting makes it again.
				if err = settingInitService.InitDefaults(ctx, db); err != nil {
					logger.Errorf("failed to initialize default settings: %v", err)
				}
			}
			// The data exists, whoever made it: what the first boot needed is
			// not kept on disk a moment longer.
			forgetFirstBootEnv(logger)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

// forgetFirstBootEnv deletes the settings the installer left for the first boot.
// A failure is logged, not fatal: the boot has what it needs, and the next boot
// tries again.
func forgetFirstBootEnv(logger logging.Logger) {
	if err := config.RemoveFirstBootEnv(); err != nil {
		logger.Errorf("failed to remove the first-boot settings: %v", err)
	}
}

func sysInstallationInitData(
	ctx context.Context,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	userService userservice.Service,
	settingInitService settinginitservice.Service,
	logger logging.Logger,
) error {
	logger.Info("initializing system data...")
	var postInitFunc func() error
	err := transaction.Execute(ctx, db, func(db database.Tx) error {
		sysStatus, err := sysStatusRepo.Get(ctx, db,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return fmt.Errorf("failed to load system status: %w", err)
		}
		config.SetInstallationStep(sysStatus.NextStep)
		if sysStatus.NextStep == "" {
			return nil
		}

		if err = userService.InitAdminUser(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize admin user: %w", err)
		}

		if err = settingInitService.InitDefaultsWithTx(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize default settings: %w", err)
		}

		// Nothing is left to do once the data exists: the server has no follow-up
		// installation step.
		sysStatus.NextStep = base.InstallationStepNone
		sysStatus.UpdateVer++
		sysStatus.UpdatedAt = timeutil.NowUTC()
		err = sysStatusRepo.Upsert(ctx, db, sysStatus,
			entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
		if err != nil {
			return fmt.Errorf("failed to save system status: %w", err)
		}
		config.SetInstallationStep(sysStatus.NextStep)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to initialize system data: %w", err)
	}

	if postInitFunc != nil {
		e := postInitFunc()
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}
