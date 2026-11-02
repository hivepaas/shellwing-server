package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	swaggoGin "github.com/swaggo/gin-swagger"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/auditloghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/devhelperhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/imagehandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/orghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/sessionhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/settinghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/supporthandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/synchandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/systemhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/userhandler"
)

type HandlerRegistry struct {
	baseHandler      *handler.BaseHandler
	authHandler      *authhandler.Handler
	auditLogHandler  *auditloghandler.Handler
	devHelperHandler *devhelperhandler.Handler
	imageHandler     *imagehandler.Handler
	orgHandler       *orghandler.Handler
	sessionHandler   *sessionhandler.Handler
	settingHandler   *settinghandler.Handler
	supportHandler   *supporthandler.Handler
	syncHandler      *synchandler.Handler
	systemHandler    *systemhandler.Handler
	userHandler      *userhandler.Handler
}

func NewHandlerRegistry(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	auditLogHandler *auditloghandler.Handler,
	devHelperHandler *devhelperhandler.Handler,
	imageHandler *imagehandler.Handler,
	orgHandler *orghandler.Handler,
	sessionHandler *sessionhandler.Handler,
	settingHandler *settinghandler.Handler,
	supportHandler *supporthandler.Handler,
	syncHandler *synchandler.Handler,
	systemHandler *systemhandler.Handler,
	userHandler *userhandler.Handler,
) *HandlerRegistry {
	return &HandlerRegistry{
		baseHandler:      baseHandler,
		authHandler:      authHandler,
		auditLogHandler:  auditLogHandler,
		devHelperHandler: devHelperHandler,
		imageHandler:     imageHandler,
		orgHandler:       orgHandler,
		sessionHandler:   sessionHandler,
		settingHandler:   settingHandler,
		supportHandler:   supportHandler,
		syncHandler:      syncHandler,
		systemHandler:    systemHandler,
		userHandler:      userHandler,
	}
}

// webAppDir holds the web app's build, which the release image copies there.
const webAppDir = "./dist-dashboard"

func (s *HTTPServer) registerRoutes() {
	s.engine.GET(s.config.HTTPServer.BasePath+"/ping", routePing)
	s.engine.NoRoute(routeNotFound(s.handlerRegistry.baseHandler))

	// Swagger server
	if s.config.IsDevEnv() {
		s.engine.Use(StaticServe("/docs", localFile("./docs", false, "")))
		s.engine.GET("/swagger/*any", swaggoGin.WrapHandler(swaggoFiles.Handler,
			swaggoGin.URL("/docs/openapi/swagger.json")))
	}

	// STATIC FILES
	s.engine.Use(StaticServe(s.config.HttpPathSslAcme(),
		localFile(s.config.DataPathSslAcme().AbsPath(), false, "")))
	registerWebApp(s.engine, webAppDir)

	// INTERNAL ROUTES
	if s.config.IsDevEnv() {
		basicAuthMdlw := gin.BasicAuth(gin.Accounts{
			s.config.Session.BasicAuthUsername: s.config.Session.BasicAuthPassword,
		})
		v1BasicAuth := s.engine.Group(s.config.HTTPServer.BasePath + "/internal")
		v1BasicAuth.Use(basicAuthMdlw)
		s.registerDevRoutes(v1BasicAuth)
	}

	// PUBLIC ROUTES
	s.registerAPIRoutes(s.engine.Group(s.config.HTTPServer.BasePath))
}

// registerAPIRoutes registers the API's routes, every one the spec documents. The route-versus-spec test registers
// them through this too, so a route added here is one it checks.
func (s *HTTPServer) registerAPIRoutes(apiGroup *gin.RouterGroup) {
	s.registerSessionRoutes(apiGroup)
	s.registerUserRoutes(apiGroup)
	s.registerSettingRoutes(apiGroup)
	s.registerSystemRoutes(apiGroup)
	s.registerImageRoutes(apiGroup)
	s.registerSupportRoutes(apiGroup)
	s.registerSyncRoutes(apiGroup)
	s.registerOrgRoutes(apiGroup)
}

// registerWebApp serves the web app at the root: its files as they are, with no directory listing, and its index.html
// for any other page, which the web app routes itself (WebAppFallback).
func registerWebApp(engine *gin.Engine, root string) {
	engine.Use(StaticServe("/", localFile(root, false, "")))
	engine.Use(WebAppFallback(root))
}

func routePing(c *gin.Context) {
	c.String(http.StatusOK, "{\"message\": \"pong\"}")
}

// routeNotFound answers what nothing else did - an API path no route has, a missing asset, a write outside the API -
// with the API's error body, so that a client of the API reads a 404 as it reads any other error.
func routeNotFound(baseHandler *handler.BaseHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseHandler.RenderError(c, hperrors.NewNotFound("Path"))
	}
}
