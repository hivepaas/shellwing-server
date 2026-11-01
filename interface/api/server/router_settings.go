package server

import (
	"github.com/gin-gonic/gin"
)

//nolint:funlen
func (s *HTTPServer) registerSettingRoutes(apiGroup *gin.RouterGroup) {
	settingGroup := apiGroup.Group("/settings")
	settingHandler := s.handlerRegistry.settingHandler

	{ // key auth group
		keyAuthGroup := settingGroup.Group("/key-auth")
		keyAuthGroup.GET("/:itemID", settingHandler.GetKeyAuth)
		keyAuthGroup.GET("", settingHandler.ListKeyAuth)
		keyAuthGroup.POST("", settingHandler.CreateKeyAuth)
		keyAuthGroup.PUT("/:itemID", settingHandler.UpdateKeyAuth)
		keyAuthGroup.PUT("/:itemID/status", settingHandler.UpdateKeyAuthStatus)
		keyAuthGroup.DELETE("/:itemID", settingHandler.DeleteKeyAuth)
	}

	{ // Cloud storage group
		cloudStorageGroup := settingGroup.Group("/cloud-storages")
		cloudStorageGroup.GET("/:itemID", settingHandler.GetCloudStorage)
		cloudStorageGroup.GET("", settingHandler.ListCloudStorage)
		cloudStorageGroup.POST("", settingHandler.CreateCloudStorage)
		cloudStorageGroup.PUT("/:itemID", settingHandler.UpdateCloudStorage)
		cloudStorageGroup.PUT("/:itemID/status", settingHandler.UpdateCloudStorageStatus)
		cloudStorageGroup.DELETE("/:itemID", settingHandler.DeleteCloudStorage)
		// Test connection
		cloudStorageGroup.POST("/test-conn", settingHandler.TestCloudStorageConn)
	}

	{ // email group
		emailGroup := settingGroup.Group("/emails")
		emailGroup.GET("/:itemID", settingHandler.GetEmail)
		emailGroup.GET("", settingHandler.ListEmail)
		emailGroup.POST("", settingHandler.CreateEmail)
		emailGroup.PUT("/:itemID", settingHandler.UpdateEmail)
		emailGroup.PUT("/:itemID/status", settingHandler.UpdateEmailStatus)
		emailGroup.DELETE("/:itemID", settingHandler.DeleteEmail)
		// Test connection
		emailGroup.POST("/test-send-mail", settingHandler.TestSendMail)
	}

	{ // IM service group
		imServiceGroup := settingGroup.Group("/im-services")
		imServiceGroup.GET("/:itemID", settingHandler.GetIMService)
		imServiceGroup.GET("", settingHandler.ListIMService)
		imServiceGroup.POST("", settingHandler.CreateIMService)
		imServiceGroup.PUT("/:itemID", settingHandler.UpdateIMService)
		imServiceGroup.PUT("/:itemID/status", settingHandler.UpdateIMServiceStatus)
		imServiceGroup.DELETE("/:itemID", settingHandler.DeleteIMService)
		// Test connection
		imServiceGroup.POST("/test-send-msg", settingHandler.TestSendInstantMsg)
	}

	{ // oauth group
		oauthGroup := settingGroup.Group("/oauth")
		oauthGroup.GET("/:itemID", settingHandler.GetOAuth)
		oauthGroup.GET("", settingHandler.ListOAuth)
		oauthGroup.POST("", settingHandler.CreateOAuth)
		oauthGroup.PUT("/:itemID", settingHandler.UpdateOAuth)
		oauthGroup.PUT("/:itemID/status", settingHandler.UpdateOAuthStatus)
		oauthGroup.DELETE("/:itemID", settingHandler.DeleteOAuth)
	}

	{ // notification group
		notificationGroup := settingGroup.Group("/notifications")
		notificationGroup.GET("/:itemID", settingHandler.GetNotification)
		notificationGroup.GET("", settingHandler.ListNotification)
		notificationGroup.POST("", settingHandler.CreateNotification)
		notificationGroup.PUT("/:itemID", settingHandler.UpdateNotification)
		notificationGroup.PUT("/:itemID/status", settingHandler.UpdateNotificationStatus)
		notificationGroup.DELETE("/:itemID", settingHandler.DeleteNotification)
	}
}
