package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerOrgRoutes(apiGroup *gin.RouterGroup) {
	orgHandler := s.handlerRegistry.orgHandler
	orgGroup := apiGroup.Group("/orgs")

	orgGroup.GET("", orgHandler.ListOrgs)
	orgGroup.GET("/:orgID/members", orgHandler.ListMembers)
	orgGroup.PUT("/:orgID/members/:userID/role", orgHandler.UpdateMemberRole)
	orgGroup.PUT("/:orgID/members/:userID/status", orgHandler.UpdateMemberStatus)

	groupGroup := orgGroup.Group("/:orgID/groups")
	groupGroup.GET("", orgHandler.ListGroups)
	groupGroup.POST("", orgHandler.CreateGroup)
	groupGroup.PUT("/:groupID", orgHandler.UpdateGroup)
	groupGroup.DELETE("/:groupID", orgHandler.DeleteGroup)
	groupGroup.PUT("/:groupID/members", orgHandler.SetGroupMembers)

	sshGroup := orgGroup.Group("/:orgID/ssh")
	sshGroup.GET("/cas", orgHandler.ListSSHCAs)
	sshGroup.POST("/cas", orgHandler.CreateSSHCA)
	sshGroup.DELETE("/cas/:caID", orgHandler.DeleteSSHCA)
	sshGroup.PUT("/settings", orgHandler.UpdateSSHSettings)
	sshGroup.GET("/members/:userID/principals", orgHandler.GetSSHPrincipals)
	sshGroup.PUT("/members/:userID/principals", orgHandler.SetSSHPrincipals)
	sshGroup.POST("/certificates", orgHandler.IssueSSHCertificate)
}
