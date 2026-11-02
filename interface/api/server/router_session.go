package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSessionRoutes(apiGroup *gin.RouterGroup) {
	sessionGroup := apiGroup.Group("/sessions")
	sessionHandler := s.handlerRegistry.sessionHandler

	// User info
	sessionGroup.GET("/me", sessionHandler.GetMe)
	// Session handling
	sessionGroup.POST("/refresh", sessionHandler.RefreshSession)
	sessionGroup.DELETE("", sessionHandler.DeleteSession)
	sessionGroup.POST("/delete-all", sessionHandler.DeleteAllSessions)

	// Apps on people's devices
	deviceGroup := apiGroup.Group("/devices")
	deviceGroup.GET("", sessionHandler.ListDevices)
	deviceGroup.DELETE("/:deviceID", sessionHandler.RevokeDevice)
	deviceGroup.POST("/authorize", sessionHandler.AuthorizeDevice)
	deviceGroup.POST("/token", sessionHandler.ExchangeDeviceCode)
	deviceGroup.POST("/refresh", sessionHandler.RefreshDevice)
	deviceGroup.POST("/sign-out", sessionHandler.SignOutDevice)

	authGroup := apiGroup.Group("/auth")
	{
		// Login options
		authGroup.GET("/login-options", sessionHandler.LoginGetOptions)
		// Sign-up options
		authGroup.GET("/signup-options", sessionHandler.SignupGetOptions)
		// Login with password
		authGroup.POST("/login-with-password", sessionHandler.LoginWithPassword)
		authGroup.POST("/login-with-passcode", sessionHandler.LoginWithPasscode)
		// Login via SSO
		authGroup.GET("/sso/:provider", sessionHandler.SSOOAuthBegin)
		authGroup.GET("/sso/callback/:provider", sessionHandler.SSOOAuthCallback)
		authGroup.POST("/sso/callback/:provider", sessionHandler.SSOOAuthCallbackPost)
		// Sign-up of somebody a provider's sign-in found no account for
		authGroup.POST("/sso/signup-info", sessionHandler.GetSSOSignupInfo)
		authGroup.POST("/sso/signup", sessionHandler.CompleteSSOSignup)
		// Password forgot
		authGroup.POST("/login-password-forgot", sessionHandler.LoginPasswordForgot)
	}
}
