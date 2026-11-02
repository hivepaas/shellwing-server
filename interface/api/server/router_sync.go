package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSyncRoutes(apiGroup *gin.RouterGroup) {
	syncHandler := s.handlerRegistry.syncHandler
	syncGroup := apiGroup.Group("/sync")

	syncGroup.GET("/spaces", syncHandler.ListSyncSpaces)

	spaceGroup := syncGroup.Group("/spaces/:spaceID")
	spaceGroup.GET("/changes", syncHandler.ListSyncChanges)
	spaceGroup.GET("/events", syncHandler.StreamSyncEvents)
	spaceGroup.GET("/object", syncHandler.GetSyncObject)
	spaceGroup.PUT("/object", syncHandler.PutSyncObject)
	spaceGroup.DELETE("/object", syncHandler.DeleteSyncObject)
}
