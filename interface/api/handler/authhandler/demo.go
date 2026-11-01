package authhandler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
)

// The demo user is a public account: it reads, and does nothing else.
//
// Its session is limited to READ, but that limit only holds where a handler
// asks for an access check, and asks for the right one. Here it holds for every
// request, by its method, wherever it is authenticated, before any handler can
// fall back on the caller's own user.

// demoWrites are the requests other than reads the demo user may make, by
// method and route suffix: they change nothing a visitor shares.
var demoWrites = []string{
	http.MethodPost + " /sessions/refresh",
	http.MethodDelete + " /sessions",
}

// DemoAllows says whether the demo user may make a request, by its method and
// the route it matched (gin's FullPath).
func DemoAllows(method, route string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		for _, write := range demoWrites {
			allowedMethod, suffix, _ := strings.Cut(write, " ")
			if method == allowedMethod && strings.HasSuffix(route, suffix) {
				return true
			}
		}
		return false
	}
}

// refuseDemo refuses a request of the demo user that DemoAllows does not allow.
func refuseDemo(ctx *gin.Context, user *basedto.User) error {
	if user == nil || user.User == nil || !user.IsDemoUser() {
		return nil
	}
	if DemoAllows(ctx.Request.Method, ctx.FullPath()) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
}
