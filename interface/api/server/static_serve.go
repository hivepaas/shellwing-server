package server

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/config"
)

// isAPIPath is whether a request path is the API's, under its configured base
// path: never a static file, and never sent to the dashboard's index.
func isAPIPath(requestPath string) bool {
	return strings.HasPrefix(requestPath, strings.TrimSuffix(config.Current().HTTPServer.BasePath, "/")+"/")
}

const INDEX = "index.html"

const (
	// webAppAssetsPrefix is where the web app's build puts its scripts, styles and images, each named by its content's
	// hash: a name there never changes what it holds, so it is cached for good.
	webAppAssetsPrefix = "/assets/"
	// assetCacheControl is for a hashed asset; htmlCacheControl for the web app's HTML, which must never be cached:
	// it names the assets of the build that is running, and a stale copy asks for ones that are gone.
	assetCacheControl = "public, max-age=31536000, immutable"
	htmlCacheControl  = "no-cache, no-store, must-revalidate"
)

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
	CacheControlHeader() string
}

// StaticServe returns a middleware handler that serves static files in the given directory.
func StaticServe(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if fs.Exists(urlPrefix, reqPath) {
			if strings.HasPrefix(reqPath, webAppAssetsPrefix) {
				c.Writer.Header().Set("Cache-Control", assetCacheControl)
			} else if strings.HasSuffix(reqPath, ".html") || reqPath == "/" {
				c.Writer.Header().Set("Cache-Control", htmlCacheControl)
			} else if cacheHeader := fs.CacheControlHeader(); cacheHeader != "" {
				c.Writer.Header().Set("Cache-Control", cacheHeader)
			}
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

type localFileSystem struct {
	http.FileSystem
	root         string
	indexes      bool
	cacheControl string
}

func localFile(root string, indexes bool, cacheControl string) *localFileSystem {
	return &localFileSystem{
		FileSystem:   gin.Dir(root, indexes),
		root:         root,
		indexes:      indexes,
		cacheControl: cacheControl,
	}
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	if isAPIPath(filepath) {
		return false
	}
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (l *localFileSystem) CacheControlHeader() string {
	return l.cacheControl
}

// WebAppFallback answers a page of the web app that is no file with the web app's index.html, the URL left as it is:
// the web app routes a page in the browser, so /pricing?installationId=x&period=yearly - a link shellwing sends a
// visitor who is not signed in - opens the pricing page with both parameters, and a link in an email opens its page.
//
// It answers only what can be a page: a GET or HEAD, outside the API, and neither under the web app's assets nor a
// file name with an extension. A missing asset is a 404, not HTML a script or image tag cannot use; the API answers
// its own 404. Without a web build in root, as while developing with the web app served on its own port, it answers
// nothing.
func WebAppFallback(root string) gin.HandlerFunc {
	index := filepath.Join(root, INDEX)
	return func(c *gin.Context) {
		if !isWebAppPage(c.Request) {
			return
		}
		file, err := os.Open(index)
		if err != nil {
			return
		}
		defer func() { _ = file.Close() }()
		stat, err := file.Stat()
		if err != nil || stat.IsDir() {
			return
		}
		c.Writer.Header().Set("Cache-Control", htmlCacheControl)
		// ServeContent answers a HEAD and a conditional GET as for the file itself, and never redirects: the URL is
		// the page's, and stays.
		http.ServeContent(c.Writer, c.Request, INDEX, stat.ModTime(), file)
		c.Abort()
	}
}

// isWebAppPage is whether a request can be for a page of the web app, which has no file of its own.
func isWebAppPage(req *http.Request) bool {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	reqPath := req.URL.Path
	if isAPIPath(reqPath) || strings.HasPrefix(reqPath, webAppAssetsPrefix) ||
		reqPath == strings.TrimSuffix(webAppAssetsPrefix, "/") {
		return false
	}
	// The web app's pages are words and IDs, never names with an extension.
	return path.Ext(reqPath) == ""
}
