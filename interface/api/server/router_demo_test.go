package server

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/interface/api/handler/auditloghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/imagehandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/sessionhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/settinghandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/supporthandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/systemhandler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/userhandler"
)

// allRoutes is every route the API registers, with the handler it calls: the
// handlers are empty, routing never calls them here. The dev helper's are left
// out; they need a server to be built.
func allRoutes(t *testing.T) gin.RoutesInfo {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	s := &HTTPServer{handlerRegistry: &HandlerRegistry{
		authHandler:     &authhandler.Handler{},
		auditLogHandler: &auditloghandler.Handler{},
		imageHandler:    &imagehandler.Handler{},
		sessionHandler:  &sessionhandler.Handler{},
		settingHandler:  &settinghandler.Handler{},
		supportHandler:  &supporthandler.Handler{},
		systemHandler:   &systemhandler.Handler{},
		userHandler:     &userhandler.Handler{},
	}}
	s.registerAPIRoutes(engine.Group("/api"))
	return engine.Routes()
}

// The demo user writes nothing but its session, across every route the API
// has.
func TestTheDemoUserWritesNothingElse(t *testing.T) {
	var allowed []string
	for _, route := range allRoutes(t) {
		if route.Method != http.MethodGet && authhandler.DemoAllows(route.Method, route.Path) {
			allowed = append(allowed, route.Method+" "+route.Path)
		}
	}

	assert.ElementsMatch(t, []string{
		"POST /api/sessions/refresh",
		"DELETE /api/sessions",
	}, allowed)
}
