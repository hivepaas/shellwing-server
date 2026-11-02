package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerOrgRoutes(apiGroup *gin.RouterGroup) {
	orgHandler := s.handlerRegistry.orgHandler
	orgGroup := apiGroup.Group("/orgs")

	orgGroup.GET("", orgHandler.ListOrgs)

	sshGroup := orgGroup.Group("/:orgID/ssh")
	sshGroup.GET("/cas", orgHandler.ListSSHCAs)
	sshGroup.POST("/cas", orgHandler.CreateSSHCA)
	sshGroup.DELETE("/cas/:caID", orgHandler.DeleteSSHCA)
	sshGroup.PUT("/settings", orgHandler.UpdateSSHSettings)
	sshGroup.GET("/members/:userID/principals", orgHandler.GetSSHPrincipals)
	sshGroup.PUT("/members/:userID/principals", orgHandler.SetSSHPrincipals)
	sshGroup.POST("/certificates", orgHandler.IssueSSHCertificate)
}
