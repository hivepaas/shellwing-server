package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerIdentityRoutes(apiGroup *gin.RouterGroup) {
	identityHandler := s.handlerRegistry.identityHandler
	apiGroup.GET("/users/current/identity", identityHandler.GetMine)
	apiGroup.PUT("/users/current/identity", identityHandler.PutMine)
}
