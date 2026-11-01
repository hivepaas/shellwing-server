package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
)

func (cfg *Config) BaseAPIURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL(), cfg.HTTPServer.BasePath))
}

func (cfg *Config) BaseDashboardURL() string {
	if cfg.WebBaseURL != "" {
		return cfg.WebBaseURL
	}
	return cfg.BaseURL()
}

var ErrWebBaseURLInvalid = errors.New("web base URL is invalid")

// validateWebBaseURL accepts an empty URL, which means the web app is served beside the API, or an
// absolute http(s) URL that every link can be joined onto: no query and no fragment.
func validateWebBaseURL(webBaseURL string) error {
	if webBaseURL == "" {
		return nil
	}
	u, err := url.Parse(webBaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("%w: SHW_WEB_BASE_URL %q - an http or https URL such as http://localhost:4100",
			ErrWebBaseURLInvalid, webBaseURL)
	}
	return nil
}

/// FRONT-END DASHBOARD

// Users

func (cfg *Config) DashboardSsoSuccessURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/sso/success"))
}

// DashboardSsoErrorURL is the sign-in page, told why a sign-in with a provider failed: the error's code, which the
// page puts in words.
func (cfg *Config) DashboardSsoErrorURL(code string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/sign-in")) + "?ssoError=" + url.QueryEscape(code)
}

func (cfg *Config) DashboardUserSignupURL(token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/sign-up")) +
		fmt.Sprintf("?token=%s", token)
}

func (cfg *Config) DashboardPasswordResetURL(userID, token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/reset-password")) +
		fmt.Sprintf("?userID=%s&token=%s", userID, token)
}

// Tasks

func (cfg *Config) DashboardTaskDetailsURL(taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "tasks", taskID))
}

/// BACK-END

func (cfg *Config) SsoBaseCallbackURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "auth/sso/callback"))
}

// SsoAuthPath is where a login with a provider begins, as the dashboard links it.
func (cfg *Config) SsoAuthPath(id string) string {
	return gofn.Must(url.JoinPath(cfg.HTTPServer.BasePath, "auth/sso", id))
}

func (cfg *Config) SsoCallbackURL(id string) string {
	return gofn.Must(url.JoinPath(cfg.SsoBaseCallbackURL(), id))
}

/// LOCAL PATH

type LocalPath string

func (lp LocalPath) RelPath() string {
	return string(lp)
}
func (lp LocalPath) AbsPath() string {
	return filepath.Join(Current().AppPath, string(lp))
}
func (lp LocalPath) Join(elem ...string) LocalPath {
	return LocalPath(filepath.Join(append([]string{string(lp)}, elem...)...))
}

/// SSL CERTS

func (cfg *Config) DataPathSsl() LocalPath {
	return "ssl"
}
func (cfg *Config) DataPathSslCerts() LocalPath {
	return cfg.DataPathSsl().Join("certs")
}
func (cfg *Config) DataPathSslAcme() LocalPath {
	return cfg.DataPathSsl().Join("acme")
}
func (cfg *Config) HttpPathSslAcme() string {
	return "/acme/"
}

/// TRAEFIK

func (cfg *Config) DataPathTraefik() LocalPath {
	return "traefik"
}
func (cfg *Config) DataPathTraefikEtc() LocalPath {
	return cfg.DataPathTraefik().Join("etc")
}
func (cfg *Config) DataPathTraefikEtcDynamic() LocalPath {
	return cfg.DataPathTraefikEtc().Join("dynamic")
}

/// UPLOAD FILES

func (cfg *Config) DataPathFiles() LocalPath {
	return "files"
}

/// DIRS TO CREATE AT STARTUP

func (cfg *Config) DataPathsToInitAtStartup() map[string]os.FileMode {
	return map[string]os.FileMode{
		cfg.DataPathSslCerts().AbsPath(): base.DirModeDefault,
		cfg.DataPathSslAcme().AbsPath():  base.DirModeDefault,

		cfg.DataPathTraefikEtcDynamic().AbsPath(): base.DirModeDefault,

		cfg.DataPathFiles().AbsPath(): base.DirModeDefault,
	}
}
