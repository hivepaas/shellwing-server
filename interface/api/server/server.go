package server

import (
	"context"
	"net/http"
	"time"

	ginlogger "github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/middleware/cors"
	loggermiddleware "github.com/hivepaas/shellwing-server/interface/api/middleware/logger"
	"github.com/hivepaas/shellwing-server/interface/api/middleware/recovery"
	"github.com/hivepaas/shellwing-server/interface/api/middleware/secretguard"
	"github.com/hivepaas/shellwing-server/interface/api/middleware/secureheaders"
	"github.com/hivepaas/shellwing-server/pkg/logging"
)

type Server interface {
	Start() error
	Stop(context.Context) error
	GetAddress() string
}

type HTTPServer struct {
	*http.Server
	config          *config.Config
	engine          *gin.Engine
	handlerRegistry *HandlerRegistry
	logger          logging.Logger
}

// NewHTTPServer Create new shellwing server app
//
// The annotations below are the API's own part of the OpenAPI spec: `make gen-swag`.
// A request authenticates with a session's access token, which the @security line
// says. The security definitions come last: swag reads a definition's lines up to
// the next definition.
//
// @title shellwing server API
// @version 0.1
// @description The REST API of shellwing server, the one its dashboard uses. Every path is under the
// @description install's API base path, `/api`: `GET https://<your server's domain>/api/users` lists the users.
// @description
// @description ## Authentication
// @description
// @description A request authenticates with **a session's access token**, in the header
// @description `Authorization: Bearer <access token>`. A login such as `POST /auth/login-with-password`
// @description gives it, and it expires within minutes: `POST /sessions/refresh` renews it.
// @description
// @description A request without one is answered `401`. The logins, the sign-up and password reset
// @description steps, and the public images need none.
// @description
// @description ## Errors
// @description
// @description A request that fails is answered with a 4xx or 5xx status and an `hperrors.ErrorInfo`
// @description body, whose `code` names the error, such as `ERR_NO_SESSION`.
// @contact.name shellwing
// @contact.url https://github.com/hivepaas/shellwing-server/issues
// @license.name AGPL-3.0
// @license.url https://github.com/hivepaas/shellwing-server/blob/main/LICENSE
// @BasePath /api
//
// @security BearerToken
//
// @tag.name Sessions
// @tag.description Logging in and out, and the session a login opens.
// @tag.name Sync
// @tag.description Syncing a person's encrypted files between their devices.
// @tag.name Users
// @tag.description The server's users: invitations, sign-up, profiles, passwords and two-factor authentication.
// @tag.name Global settings
// @tag.description The server's settings: key auths, storages, emails, IM services, OAuth and notifications.
// @tag.name System tasks
// @tag.description The background tasks run for the server.
// @tag.name System audit logs
// @tag.description What was done across the server, and by whom.
// @tag.name System errors
// @tag.description Errors the server ran into in the background.
// @tag.name Images
// @tag.description The public images, such as users' photos.
// @tag.name Support
// @tag.description Sending feedback to the shellwing team.
//
// @securityDefinitions.apikey BearerToken
// @in header
// @name Authorization
// @description `Bearer <access token>`, a session's access token from a login.
func NewHTTPServer(
	config *config.Config,
	logger logging.Logger,
	handlerRegistry *HandlerRegistry,
) Server {
	s := &HTTPServer{
		config:          config,
		handlerRegistry: handlerRegistry,
		logger:          logger,
	}
	return s
}

func (s *HTTPServer) init() {
	engine := gin.New()
	s.engine = engine

	applyTrustedProxies(engine, s.config.HTTPServer.TrustedProxies, s.logger)

	s.Server = &http.Server{
		Addr:           s.config.HTTPServer.BindingAddress(),
		ReadTimeout:    180 * time.Second, //nolint:mnd
		WriteTimeout:   180 * time.Second, //nolint:mnd
		MaxHeaderBytes: 1 << 20,           //nolint:mnd
		Handler:        engine,
	}

	// Configures middlewares
	engine.Use(
		recovery.Recovery(s.config, s.handlerRegistry.baseHandler),
		loggermiddleware.Logger(s.logger),
		secureheaders.SecureHeaders,
		cors.CORS(s.config),
	)

	if s.config.IsDevEnv() {
		engine.Use(ginlogger.SetLogger())
		// Development only: it copies every JSON response body, and it panics.
		// Registered after Recovery so the panic is caught and reported.
		engine.Use(secretguard.Guard())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.registerRoutes()
}

// applyTrustedProxies decides whose X-Forwarded-For header is believed.
//
// Gin's own default is to trust 0.0.0.0/0, which means any caller can pick the
// address recorded against them just by sending the header. That is harmless
// while nothing reads the address and wrong the moment something does - the audit
// log does. So nothing is trusted unless it is configured here, and a malformed
// entry falls back to trusting nothing rather than being ignored, which would
// leave gin in the very state this exists to leave.
func applyTrustedProxies(engine *gin.Engine, trustedProxies []string, logger logging.Logger) {
	if err := engine.SetTrustedProxies(trustedProxies); err == nil {
		return
	} else if logger != nil {
		logger.Errorf("invalid http_server.trusted_proxies %v: %v - trusting no proxy",
			trustedProxies, err)
	}
	_ = engine.SetTrustedProxies(nil)
}

func (s *HTTPServer) Start() error {
	if s.Server == nil {
		s.init()
	}

	err := s.ListenAndServe()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}
	err := s.Shutdown(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *HTTPServer) GetAddress() string {
	return s.Addr
}
