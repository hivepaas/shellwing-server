package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSystemRoutes(apiGroup *gin.RouterGroup) {
	systemGroup := apiGroup.Group("/system")
	systemHandler := s.handlerRegistry.systemHandler

	{ // Task group
		taskGroup := systemGroup.Group("/tasks")
		taskGroup.GET("", systemHandler.ListTask)
		taskGroup.GET("/:itemID", systemHandler.GetTask)
		taskGroup.GET("/:itemID/status", systemHandler.GetTaskStatus)
		taskGroup.POST("/:itemID/cancel", systemHandler.CancelTask)
		taskGroup.GET("/:itemID/logs", systemHandler.GetTaskLogs)
	}

	{ // Error group
		errorGroup := systemGroup.Group("/errors")
		errorGroup.GET("", systemHandler.ListSysError)
		errorGroup.GET("/:errorID", systemHandler.GetSysError)
		errorGroup.DELETE("/:errorID", systemHandler.DeleteSysError)
	}

	{ // Audit log group
		auditLogGroup := systemGroup.Group("/audit-logs")
		auditLogGroup.GET("", systemHandler.ListAuditLog)
		auditLogGroup.GET("/types", systemHandler.ListAuditLogTypes)
		auditLogGroup.GET("/:itemID", systemHandler.GetAuditLog)
	}
}
