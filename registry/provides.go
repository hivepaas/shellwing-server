package registry

import (
	"context"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/auditloghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/devhelperhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/imagehandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/orghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/sessionhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/settinghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/supporthandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/synchandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/systemhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/taskhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/userhandler"
	"github.com/hivepaas/shellwing-server/interface/api/server"
	"github.com/hivepaas/shellwing-server/permission/permissionimpl"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/auditservice/auditserviceimpl"
	"github.com/hivepaas/shellwing-server/service/datakeyservice/datakeyserviceimpl"
	"github.com/hivepaas/shellwing-server/service/emailservice/emailserviceimpl"
	"github.com/hivepaas/shellwing-server/service/licenseservice/licenseserviceimpl"
	"github.com/hivepaas/shellwing-server/service/notificationservice/notificationserviceimpl"
	"github.com/hivepaas/shellwing-server/service/orgservice/orgserviceimpl"
	"github.com/hivepaas/shellwing-server/service/reslinkservice/reslinkserviceimpl"
	"github.com/hivepaas/shellwing-server/service/scopeservice/scopeserviceimpl"
	"github.com/hivepaas/shellwing-server/service/settingeventservice/settingeventserviceimpl"
	"github.com/hivepaas/shellwing-server/service/settinginitservice/settinginitserviceimpl"
	"github.com/hivepaas/shellwing-server/service/settingservice/settingserviceimpl"
	"github.com/hivepaas/shellwing-server/service/startupservice/startupserviceimpl"
	"github.com/hivepaas/shellwing-server/service/syncservice/syncserviceimpl"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice/systemeventbusserviceimpl"
	"github.com/hivepaas/shellwing-server/service/taskservice/taskserviceimpl"
	"github.com/hivepaas/shellwing-server/service/userservice/userserviceimpl"
	"github.com/hivepaas/shellwing-server/tasks/initializer"
	"github.com/hivepaas/shellwing-server/tasks/queue/queueimpl"
	"github.com/hivepaas/shellwing-server/tasks/taskdummy"
	"github.com/hivepaas/shellwing-server/usecase/auditloguc"
	"github.com/hivepaas/shellwing-server/usecase/binobjectuc"
	"github.com/hivepaas/shellwing-server/usecase/devhelperuc"
	"github.com/hivepaas/shellwing-server/usecase/orguc"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc"
	"github.com/hivepaas/shellwing-server/usecase/settings"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc"
	"github.com/hivepaas/shellwing-server/usecase/supportuc"
	"github.com/hivepaas/shellwing-server/usecase/syncuc"
	"github.com/hivepaas/shellwing-server/usecase/system/licenseuc"
	"github.com/hivepaas/shellwing-server/usecase/system/syserroruc"
	"github.com/hivepaas/shellwing-server/usecase/taskuc"
	"github.com/hivepaas/shellwing-server/usecase/useruc"
)

var Provides = []any{
	context.Background,

	// Configuration
	config.LoadConfig,

	// Logger
	logging.NewZapLogger,

	// DB
	database.NewDB,

	// Cache
	rediscache.NewClient,
	rediscache.NewLock,

	// HTTP server
	server.NewHTTPServer,

	// Permission
	permissionimpl.NewManager,

	// Task queue
	initializer.NewWorkerInitializer,
	queueimpl.New,
	taskdummy.NewExecutor,

	// Route handler
	auditloghandler.New,
	authhandler.New,
	basesettinghandler.New,
	devhelperhandler.New,
	handler.New,
	imagehandler.New,
	orghandler.New,
	server.NewHandlerRegistry,
	sessionhandler.New,
	settinghandler.New,
	supporthandler.New,
	synchandler.New,
	systemhandler.New,
	taskhandler.New,
	userhandler.New,

	// Use case
	auditloguc.New,
	binobjectuc.New,
	cloudstorageuc.New,
	devhelperuc.New,
	emailuc.New,
	imserviceuc.New,
	keyauthuc.New,
	notificationuc.New,
	oauthuc.New,
	sessionuc.New,
	settings.New,
	supportuc.New,
	orguc.New,
	syncuc.New,
	licenseuc.New,
	syserroruc.New,
	taskuc.New,
	useruc.New,

	// Service
	auditserviceimpl.New,
	datakeyserviceimpl.New,
	emailserviceimpl.New,
	licenseserviceimpl.New,
	notificationserviceimpl.New,
	orgserviceimpl.New,
	reslinkserviceimpl.New,
	scopeserviceimpl.New,
	settingeventserviceimpl.New,
	settinginitserviceimpl.New,
	settingserviceimpl.New,
	startupserviceimpl.New,
	syncserviceimpl.New,
	syncserviceimpl.NewStore,
	systemeventbusserviceimpl.New,
	taskserviceimpl.New,
	userserviceimpl.New,

	// Repository
	repository.NewACLPermissionRepo,
	repository.NewAuditLogRepo,
	repository.NewBinObjectRepo,
	repository.NewDataMigrationRepo,
	repository.NewDeviceRepo,
	repository.NewEncryptionKeyRepo,
	repository.NewFileRepo,
	repository.NewLicenseRepo,
	repository.NewLockRepo,
	repository.NewLoginTrustedDeviceRepo,
	repository.NewOrganizationRepo,
	repository.NewSSHCARepo,
	repository.NewOrgGroupRepo,
	repository.NewResLinkRepo,
	repository.NewSettingRepo,
	repository.NewSpaceRepo,
	repository.NewSyncRepo,
	repository.NewSysErrorRepo,
	repository.NewSystemStatusRepo,
	repository.NewTagRepo,
	repository.NewTaskLogRepo,
	repository.NewTaskRepo,
	repository.NewUserRepo,

	// Cache repository
	cacherepository.NewDeviceAuthCodeRepo,
	cacherepository.NewLoginAttemptRepo,
	cacherepository.NewMFAPasscodeRepo,
	cacherepository.NewRateLimitRepo,
	cacherepository.NewTaskControlRepo,
	cacherepository.NewTaskInfoRepo,
	cacherepository.NewUserTokenRepo,
}
