package orghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/basedto"
	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/orguc"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

const (
	paramOrgID   = "orgID"
	paramCAID    = "caID"
	paramUserID  = "userID"
	paramGroupID = "groupID"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	orgUC       *orguc.UC
}

func New(baseHandler *handler.BaseHandler, authHandler *authhandler.Handler, orgUC *orguc.UC) *Handler {
	return &Handler{BaseHandler: baseHandler, authHandler: authHandler, orgUC: orgUC}
}

// adminAuth is a person signed in, for what an organization's admin does. An app's token does not reach it: the role in
// the organization is the use case's to check.
func (h *Handler) adminAuth(ctx *gin.Context) (*basedto.Auth, error) {
	return h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck) //nolint:wrapcheck // the handler's own
}

// ListOrgs lists the person's organizations
// @Summary Lists the person's organizations
// @Description The organizations the person is an active member of, with their role and the accounts their SSH
// @Description certificates may log in as.
// @Tags    Organizations
// @Produce json
// @Id      listOrgs
// @Success 200 {object} orgdto.ListOrgsResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs [get]
func (h *Handler) ListOrgs(ctx *gin.Context) {
	auth, err := h.authHandler.GetSyncAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.ListOrgs(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ListSSHCAs lists the organization's certificate authorities
// @Summary Lists the organization's SSH certificate authorities
// @Description Their public keys, for sshd's TrustedUserCAKeys. A retired one signs no more, and is listed until
// @Description deleted so that servers trusting it can be moved over.
// @Tags    Organizations
// @Produce json
// @Id      listSSHCAs
// @Param   orgID path string true "organization ID"
// @Success 200 {object} orgdto.ListSSHCAsResp
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/cas [get]
func (h *Handler) ListSSHCAs(ctx *gin.Context) {
	auth, err := h.authHandler.GetSyncAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.ListSSHCAs(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// CreateSSHCA makes a new certificate authority
// @Summary Makes a new SSH certificate authority
// @Description The new one signs from now on. One before it is retired and its private key destroyed. Certificates
// @Description signed from now on work only on servers that trust the new one.
// @Tags    Organizations
// @Produce json
// @Id      createSSHCA
// @Param   orgID path string true "organization ID"
// @Success 201 {object} orgdto.CreateSSHCAResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/cas [post]
func (h *Handler) CreateSSHCA(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.CreateSSHCA(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// DeleteSSHCA deletes a retired certificate authority
// @Summary Deletes a retired SSH certificate authority
// @Description The active one is replaced, not deleted.
// @Tags    Organizations
// @Produce json
// @Id      deleteSSHCA
// @Param   orgID path string true "organization ID"
// @Param   caID path string true "certificate authority ID"
// @Success 200 {object} orgdto.DeleteSSHCAResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/cas/{caID} [delete]
func (h *Handler) DeleteSSHCA(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.DeleteSSHCA(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramCAID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// UpdateSSHSettings sets how long certificates last
// @Summary Sets how long the organization's SSH certificates last
// @Description From 5 minutes to a day. Certificates signed already keep their end.
// @Tags    Organizations
// @Produce json
// @Id      updateSSHSettings
// @Param   orgID path string true "organization ID"
// @Param   body body orgdto.UpdateSSHSettingsReq true "request data"
// @Success 200 {object} orgdto.UpdateSSHSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/settings [put]
func (h *Handler) UpdateSSHSettings(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewUpdateSSHSettingsReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.UpdateSSHSettings(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ListMembers lists the organization's members
// @Summary Lists the organization's members
// @Description By username. An admin sees each member's email, status, account status and the accounts their SSH
// @Description certificates may log in as, removed members included. A member sees the active members' names and roles.
// @Tags    Organizations
// @Produce json
// @Id      listOrgMembers
// @Param   orgID path string true "organization ID"
// @Success 200 {object} orgdto.ListMembersResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/members [get]
func (h *Handler) ListMembers(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.ListMembers(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// UpdateMemberRole changes a member's role
// @Summary Changes a member's role in the organization
// @Description Owners and admins make members admins and admins members. Only an owner makes an owner or changes one,
// @Description and the last active owner stays one (ERR_ORG_LAST_OWNER).
// @Tags    Organizations
// @Produce json
// @Id      updateOrgMemberRole
// @Param   orgID path string true "organization ID"
// @Param   userID path string true "user ID"
// @Param   body body orgdto.UpdateMemberRoleReq true "request data"
// @Success 200 {object} orgdto.UpdateMemberRoleResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/members/{userID}/role [put]
func (h *Handler) UpdateMemberRole(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewUpdateMemberRoleReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.UpdateMemberRole(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramUserID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// UpdateMemberStatus suspends, removes or reactivates a member
// @Summary Suspends a member, removes them, or makes them active again
// @Description Owners and admins do it to members and admins, and only an owner to an owner. Anybody may remove
// @Description themselves: that is leaving. The last active owner stays one (ERR_ORG_LAST_OWNER). Removing a member
// @Description takes their role and SSH accounts away.
// @Tags    Organizations
// @Produce json
// @Id      updateOrgMemberStatus
// @Param   orgID path string true "organization ID"
// @Param   userID path string true "user ID"
// @Param   body body orgdto.UpdateMemberStatusReq true "request data"
// @Success 200 {object} orgdto.UpdateMemberStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/members/{userID}/status [put]
func (h *Handler) UpdateMemberStatus(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewUpdateMemberStatusReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.UpdateMemberStatus(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramUserID),
		req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetSSHPrincipals gets the accounts a member may log in as
// @Summary Gets the accounts a member's SSH certificates may log in as
// @Description What an admin granted them.
// @Tags    Organizations
// @Produce json
// @Id      getSSHPrincipals
// @Param   orgID path string true "organization ID"
// @Param   userID path string true "user ID"
// @Success 200 {object} orgdto.SSHPrincipalsResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/members/{userID}/principals [get]
func (h *Handler) GetSSHPrincipals(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.GetSSHPrincipals(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramUserID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// SetSSHPrincipals sets the accounts a member may log in as
// @Summary Sets the accounts a member's SSH certificates may log in as
// @Description Replaces them, from the next certificate on. Empty takes them all away. Nobody has any until granted.
// @Tags    Organizations
// @Produce json
// @Id      setSSHPrincipals
// @Param   orgID path string true "organization ID"
// @Param   userID path string true "user ID"
// @Param   body body orgdto.SetSSHPrincipalsReq true "request data"
// @Success 200 {object} orgdto.SSHPrincipalsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/members/{userID}/principals [put]
func (h *Handler) SetSSHPrincipals(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewSetSSHPrincipalsReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.SetSSHPrincipals(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramUserID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// IssueSSHCertificate signs the person's public key
// @Summary Signs the person's SSH public key
// @Description A short-lived certificate for the accounts an admin granted them. The private key stays on the device.
// @Description Each one is recorded in the audit log.
// @Tags    Organizations
// @Produce json
// @Id      issueSSHCertificate
// @Param   orgID path string true "organization ID"
// @Param   body body orgdto.IssueSSHCertificateReq true "request data"
// @Success 201 {object} orgdto.IssueSSHCertificateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/ssh/certificates [post]
func (h *Handler) IssueSSHCertificate(ctx *gin.Context) {
	auth, err := h.authHandler.GetSyncAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewIssueSSHCertificateReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.IssueSSHCertificate(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}
