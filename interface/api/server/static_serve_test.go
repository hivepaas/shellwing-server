package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
)

const testIndexHTML = `<!doctype html><html><body><div id="root"></div></body></html>`

// webAppEngine is the server's routing of the web app over a build in root, beside an API route, under the API's
// base path /api.
func webAppEngine(t *testing.T, root string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	previous := config.Current()
	cfg := &config.Config{}
	cfg.HTTPServer.BasePath = "/api"
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })

	engine := gin.New()
	engine.GET("/api/ping", routePing)
	engine.NoRoute(routeNotFound(&handler.BaseHandler{}))
	registerWebApp(engine, root)
	engine.GET("/api/licenses", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return engine
}

// webBuild writes a web build: its index.html and a hashed asset.
func webBuild(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, INDEX), []byte(testIndexHTML), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "index-B2x9.js"), []byte("console.log(1)"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "favicon.ico"), []byte("icon"), 0o600))
	return root
}

func serve(engine *gin.Engine, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

// A page of the web app is its index.html, at the URL asked for, so the web app routes it with every parameter: the
// link shellwing sends to upgrade works for a visitor who is not signed in.
func TestAPageOfTheWebAppIsItsIndex(t *testing.T) {
	engine := webAppEngine(t, webBuild(t))

	for _, target := range []string{
		"/pricing?installationId=q9xk2m4j7n1p0r5s8t3v6w2yz&period=yearly",
		"/pricing/",
		"/auth/reset-password?userID=u1&token=abc.def",
		"/account/licenses/lic_01K7P3ZQ8M5W2E4R6T8Y0B1C3D",
		"/apiary",
		"/",
	} {
		rec := serve(engine, http.MethodGet, target)
		assert.Equal(t, http.StatusOK, rec.Code, target)
		assert.Equal(t, testIndexHTML, rec.Body.String(), target)
		assert.Equal(t, htmlCacheControl, rec.Header().Get("Cache-Control"), target)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html", target)
		assert.Empty(t, rec.Header().Get("Location"), "the URL stays the page's: %s", target)
	}

	rec := serve(engine, http.MethodHead, "/pricing?installationId=x")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}

// The web app's files are served as they are: a hashed asset cached for good.
func TestTheWebAppsFilesAreServed(t *testing.T) {
	engine := webAppEngine(t, webBuild(t))

	rec := serve(engine, http.MethodGet, "/assets/index-B2x9.js")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "console.log(1)", rec.Body.String())
	assert.Equal(t, assetCacheControl, rec.Header().Get("Cache-Control"))

	rec = serve(engine, http.MethodGet, "/favicon.ico")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "icon", rec.Body.String())
}

// What is no page is a 404 in the API's error body, never the web app's HTML: a missing asset, which a script tag
// could not use; a directory, which is never listed; an API path no route has; a write outside the API.
func TestWhatIsNoPageIsNotFound(t *testing.T) {
	engine := webAppEngine(t, webBuild(t))

	for _, req := range []struct{ method, target string }{
		{http.MethodGet, "/assets/index-old.js"},
		{http.MethodGet, "/assets/"},
		{http.MethodGet, "/assets"},
		{http.MethodGet, "/robots.txt"},
		{http.MethodGet, "/api/no-such-endpoint"},
		{http.MethodGet, "/api/licenses/x/y"},
		{http.MethodPost, "/pricing"},
	} {
		rec := serve(engine, req.method, req.target)
		name := req.method + " " + req.target
		assert.Equal(t, http.StatusNotFound, rec.Code, name)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json", name)
		var body hperrors.ErrorInfo
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), name) {
			assert.Equal(t, "ERR_NOT_FOUND", body.Code, name)
			assert.Equal(t, "Path not found", body.Detail, name)
			assert.Empty(t, body.StackTrace, name)
		}
	}

	// The API's own routes are the API's.
	assert.Equal(t, http.StatusOK, serve(engine, http.MethodGet, "/api/ping").Code)
	assert.Equal(t, http.StatusNoContent, serve(engine, http.MethodGet, "/api/licenses").Code)
}

// Without a web build - while developing, the web app is served on its own port - a page is a 404.
func TestWithoutAWebBuildAPageIsNotFound(t *testing.T) {
	engine := webAppEngine(t, t.TempDir())

	assert.Equal(t, http.StatusNotFound, serve(engine, http.MethodGet, "/pricing").Code)
	assert.Equal(t, http.StatusNotFound, serve(engine, http.MethodGet, "/").Code)
}
