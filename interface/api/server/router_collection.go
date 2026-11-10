package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerCollectionRoutes(apiGroup *gin.RouterGroup) {
	collectionHandler := s.handlerRegistry.collectionHandler
	group := apiGroup.Group("/orgs/:orgID/collections")
	group.GET("", collectionHandler.ListCollections)
	group.POST("", collectionHandler.CreateCollection)
	group.GET("/:collectionID", collectionHandler.GetCollection)
	group.PUT("/:collectionID", collectionHandler.UpdateCollection)
	group.DELETE("/:collectionID", collectionHandler.DeleteCollection)
	group.PUT("/:collectionID/grants", collectionHandler.SetCollectionGrants)
	group.GET("/:collectionID/grantable", collectionHandler.GetCollectionGrantable)
	group.GET("/:collectionID/keys", collectionHandler.GetCollectionKeys)
	group.POST("/:collectionID/keys", collectionHandler.CreateCollectionKey)
	group.POST("/:collectionID/wraps", collectionHandler.AddCollectionWraps)
}
