package sessionhandler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

// SSOOAuthBegin Starts OAuth SSO flow
// @Summary Starts OAuth SSO flow
// @Description Starts OAuth SSO flow
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthBegin
// @Param   provider path string true "provider name"
// @Param   trustedDeviceId query string false "the device to trust, so no passcode is asked on it for a while"
// @Param   next query string false "the page of the web app to open once signed in, such as /account/trial"
// @Success 302 "redirect to the provider; on a failure, to the sign-in page with ssoError, the error's code"
// @Security
// @Router  /auth/sso/{provider} [get]
func (h *Handler) SSOOAuthBegin(ctx *gin.Context) {
	next := ssoNext(ctx.Query("next"))
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.redirectSSOError(ctx, err, next)
		return
	}

	err = h.sessionUC.InitOAuthProvider(ctx, &sessiondto.InitOAuthProviderReq{Provider: provider})
	if err != nil {
		h.redirectSSOError(ctx, err, next)
		return
	}

	deviceID := strings.TrimSpace(ctx.Query("trustedDeviceId"))
	h.keepForSSOCallback(ctx, cookieSSOTrustedDevice, gofn.If(len(deviceID) <= maxTrustedDeviceIDLen, deviceID, ""))
	h.keepForSSOCallback(ctx, cookieSSONext, next)

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

// SSOOAuthCallbackPost Completes the SSO flow, for a provider that posts its answer
// @Summary Completes the SSO flow, for a provider that posts its answer
// @Description The same as the GET, for a provider that returns with a form post (response_mode=form_post).
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthCallbackPost
// @Param   provider path string true "provider name"
// @Success 302 "redirect to the dashboard; on a failure, to the sign-in page with ssoError, the error's code"
// @Security
// @Router  /auth/sso/callback/{provider} [post]
//
// A handler of its own, not a second @Router on SSOOAuthCallback: an operation
// id names one operation, and a client generated from the spec needs one each.
func (h *Handler) SSOOAuthCallbackPost(ctx *gin.Context) {
	h.SSOOAuthCallback(ctx)
}

// SSOOAuthCallback Completes the SSO flow
// @Summary Completes the SSO flow
// @Description The provider sends the browser back here once the person has signed in.
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthCallback
// @Param   provider path string true "provider name"
// @Success 302 "redirect to the dashboard; on a failure, to the sign-in page with ssoError, the error's code"
// @Security
// @Router  /auth/sso/callback/{provider} [get]
func (h *Handler) SSOOAuthCallback(ctx *gin.Context) {
	next := ssoNext(h.takeFromSSOBegin(ctx, cookieSSONext))
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.redirectSSOError(ctx, err, next)
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()

	// An unknown provider, a missing or mismatched state, a code the provider
	// refuses: the sign-in is not completed, whichever it was.
	oauthUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		h.redirectSSOError(ctx, hperrors.Wrap(hperrors.ErrUnauthorized).WithCause(err), next)
		return
	}

	// Create a new session for OAuth user in our service
	sessionReq := sessiondto.NewCreateOAuthSessionReq()
	sessionReq.User = &oauthUser
	sessionReq.TrustedDeviceID = h.takeFromSSOBegin(ctx, cookieSSOTrustedDevice)
	sessionResp, err := h.sessionUC.CreateOAuthSession(h.RequestCtx(ctx), sessionReq)
	if err != nil {
		h.redirectSSOError(ctx, err, next)
		return
	}

	successURL := withNext(config.Current().DashboardSsoSuccessURL(), next)

	// A user with an authenticator app owes the passcode: the web app's sign-in asks for it with this token.
	if sessionResp.MFAToken != "" {
		h.handOverToWeb(ctx, cookieSSOMFAToken, sessionResp.MFAToken, ssoMFATokenMaxAge)
		ctx.Redirect(http.StatusFound, successURL)
		return
	}
	// Somebody with no account: the web app's sign-up page asks them to confirm the one offered.
	if sessionResp.SignupToken != "" {
		h.handOverToWeb(ctx, cookieSSOSignupToken, sessionResp.SignupToken, ssoSignupTokenMaxAge)
		ctx.Redirect(http.StatusFound, successURL)
		return
	}

	// Write session data to request cookies as we will redirect to a front-end path
	// and passing long tokens in the URL is risky.
	h.writeSessionDataToCookies(ctx, &sessionResp.BaseCreateSessionResp, false)

	// Redirect client to front-end page
	ctx.Redirect(http.StatusFound, successURL)
}

// redirectSSOError sends the browser back to the sign-in page, which says why the sign-in failed. Every step of a
// sign-in with a provider is a page the browser opened, and an error written as JSON would leave the person on it.
func (h *Handler) redirectSSOError(ctx *gin.Context, err error, next string) {
	errInfo := h.ErrorInfoOf(ctx, err)
	ctx.Redirect(http.StatusFound, withNext(config.Current().DashboardSsoErrorURL(errInfo.Code), next))
}

const (
	// cookieSSOTrustedDevice and cookieSSONext carry the device to trust and the page to open from the start of a
	// sign-in with a provider to its end, only to the SSO endpoints.
	cookieSSOTrustedDevice = "sso_trusted_device"
	cookieSSONext          = "sso_next"
	// cookieSSOMFAToken and cookieSSOSignupToken hand the web app what the sign-in ended with instead of a session:
	// the token its passcode step signs in with, or the account its sign-up page offers. The web app reads them, as
	// it reads the access token after a sign-in with a provider.
	cookieSSOMFAToken    = "mfa_token"
	cookieSSOSignupToken = "sso_signup_token"

	// ssoRoundTripMaxAge is how long a sign-in with a provider may take, there and back.
	ssoRoundTripMaxAge = 10 * 60
	// ssoMFATokenMaxAge is how long the web app has to pick the token up. The token runs out on its own, in a
	// minute: the cookie only carries it.
	ssoMFATokenMaxAge = 2 * 60
	// ssoSignupTokenMaxAge is as long as the sign-up token holds.
	ssoSignupTokenMaxAge = 15 * 60
	// maxTrustedDeviceIDLen bounds what is kept of a device ID; one longer is not kept at all.
	maxTrustedDeviceIDLen = 200
)

func (h *Handler) ssoCookiePath() string {
	return gofn.Must(url.JoinPath(config.Current().HTTPServer.BasePath, "auth/sso"))
}

// keepForSSOCallback keeps a value of the sign-in's start for the provider's callback, out of the web app's reach.
// An empty one forgets what a sign-in given up on kept.
func (h *Handler) keepForSSOCallback(ctx *gin.Context, name, value string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(name, value, gofn.If(value == "", -1, ssoRoundTripMaxAge), h.ssoCookiePath(), "",
		!config.Current().IsDevEnv(), true)
}

// takeFromSSOBegin is a value keepForSSOCallback kept, which it forgets: it is for this sign-in only.
func (h *Handler) takeFromSSOBegin(ctx *gin.Context, name string) string {
	value, err := ctx.Cookie(name)
	if err != nil {
		return ""
	}
	ctx.SetCookie(name, "", -1, h.ssoCookiePath(), "", !config.Current().IsDevEnv(), true)
	return value
}

// handOverToWeb hands the web app a token the sign-in ended with, beside where the access token goes.
func (h *Handler) handOverToWeb(ctx *gin.Context, name, token string, maxAge int) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	baseURL := strings.SplitN(config.Current().BaseURL(), ".", 2) //nolint:mnd
	ctx.SetCookie(name, token, maxAge, "", baseURL[len(baseURL)-1], !config.Current().IsDevEnv(),
		cookieAccessHTTPOnly)
}

// ssoNext is the page to open once signed in, when it is a path on the web app: anything else is dropped.
func ssoNext(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !userdto.IsSafeNextPath(next) {
		return ""
	}
	return next
}

// withNext is link with next added to its query, when there is one.
func withNext(link, next string) string {
	if next == "" {
		return link
	}
	return link + gofn.If(strings.Contains(link, "?"), "&", "?") + "next=" + url.QueryEscape(next)
}
