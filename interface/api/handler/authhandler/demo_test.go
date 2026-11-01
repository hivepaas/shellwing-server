package authhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/permission"
)

const demoUserID = "demo-user"

func withDemoUser(t *testing.T) {
	t.Helper()
	previous := config.Current()
	cfg := &config.Config{}
	cfg.Users.Demo.UserID = demoUserID
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })
}

// signedIn is a session whose token is the user's.
type signedIn struct {
	user *basedto.User
}

func (s *signedIn) GetCurrentUserByJWT(context.Context, string) (*basedto.User, error) {
	return s.user, nil
}

func (s *signedIn) GetCurrentAuthByJWT(context.Context, string) (*basedto.Auth, error) {
	return &basedto.Auth{User: s.user}, nil
}

func (s *signedIn) VerifyAuth(context.Context, *basedto.Auth, permission.AccessCheck) error {
	return nil
}

func newSignedInHandler(userID string) *Handler {
	return &Handler{BaseHandler: &handler.BaseHandler{},
		sessionUC: &signedIn{user: &basedto.User{User: &entity.User{ID: userID}}}}
}

// call sends a request to route through a router whose handler authenticates
// it, and answers what the handler was given.
func call(t *testing.T, userID, method, route, path string) (*basedto.Auth, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := newSignedInHandler(userID)
	var auth *basedto.Auth
	var err error
	engine := gin.New()
	engine.Handle(method, route, func(ctx *gin.Context) { auth, err = h.GetCurrentAuth(ctx, NoAccessCheck) })
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer token")
	engine.ServeHTTP(httptest.NewRecorder(), req)
	return auth, err
}

// The demo user reads, and does nothing else, whatever access check a handler
// asks for - none, or one that writes under a READ.
func TestTheDemoUserOnlyReads(t *testing.T) {
	withDemoUser(t)
	const user = "/api/users/:userID"

	auth, err := call(t, demoUserID, http.MethodGet, user, "/api/users/u")
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		auth, err = call(t, demoUserID, method, user, "/api/users/u")
		assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized, method)
		assert.Nil(t, auth, "%s: no auth to fall back on", method)
	}
}

// Keeping a session going and ending it write nothing anyone shares.
func TestTheDemoUserKeepsItsSession(t *testing.T) {
	withDemoUser(t)
	for _, c := range []struct{ method, route string }{
		{http.MethodPost, "/api/sessions/refresh"},
		{http.MethodDelete, "/api/sessions"},
	} {
		_, err := call(t, demoUserID, c.method, c.route, c.route)
		assert.NoError(t, err, c.method+" "+c.route)
	}
	_, err := call(t, demoUserID, http.MethodPost, "/api/sessions/delete-all", "/api/sessions/delete-all")
	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized, "every visitor shares the demo's sessions")
}

// Nobody else is concerned.
func TestOtherUsersAreNotConcerned(t *testing.T) {
	withDemoUser(t)

	_, err := call(t, "someone", http.MethodPost, "/api/users/:userID", "/api/users/u")
	assert.NoError(t, err)
}

// GetCurrentUser, which the session and user handlers use, is held to it too.
func TestGetCurrentUserOfTheDemoUserOnlyReads(t *testing.T) {
	withDemoUser(t)
	gin.SetMode(gin.TestMode)
	h := newSignedInHandler(demoUserID)
	var user *basedto.User
	var err error
	engine := gin.New()
	engine.PUT("/api/users/:userID/password", func(ctx *gin.Context) { user, err = h.GetCurrentUser(ctx) })
	req := httptest.NewRequest(http.MethodPut, "/api/users/current/password", nil)
	req.Header.Set("Authorization", "Bearer token")

	engine.ServeHTTP(httptest.NewRecorder(), req)

	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized)
	assert.Nil(t, user)
}
