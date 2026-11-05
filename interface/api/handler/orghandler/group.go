package orghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

// ListGroups lists the organization's groups
// @Summary Lists the organization's groups
// @Description With their members and the SSH accounts each grants its members, by name. For admins.
// @Tags    Organizations
// @Produce json
// @Id      listOrgGroups
// @Param   orgID path string true "organization ID"
// @Success 200 {object} orgdto.ListGroupsResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/groups [get]
func (h *Handler) ListGroups(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.ListGroups(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// CreateGroup makes a group
// @Summary Makes a group in the organization
// @Description With no members. Its name is unique in the organization, whatever its case.
// @Tags    Organizations
// @Produce json
// @Id      createOrgGroup
// @Param   orgID path string true "organization ID"
// @Param   body body orgdto.GroupReq true "request data"
// @Success 201 {object} orgdto.GroupDataResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/groups [post]
func (h *Handler) CreateGroup(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewGroupReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.CreateGroup(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// UpdateGroup changes a group
// @Summary Changes a group's name, description and SSH accounts
// @Description Its members' next certificates follow.
// @Tags    Organizations
// @Produce json
// @Id      updateOrgGroup
// @Param   orgID path string true "organization ID"
// @Param   groupID path string true "group ID"
// @Param   body body orgdto.GroupReq true "request data"
// @Success 200 {object} orgdto.GroupDataResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/groups/{groupID} [put]
func (h *Handler) UpdateGroup(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewGroupReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.UpdateGroup(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramGroupID), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DeleteGroup deletes a group
// @Summary Deletes a group
// @Description Its members lose what it granted, from their next certificate on.
// @Tags    Organizations
// @Produce json
// @Id      deleteOrgGroup
// @Param   orgID path string true "organization ID"
// @Param   groupID path string true "group ID"
// @Success 200 {object} orgdto.DeleteGroupResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/groups/{groupID} [delete]
func (h *Handler) DeleteGroup(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.DeleteGroup(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramGroupID))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// SetGroupMembers sets who is in a group
// @Summary Sets who is in a group
// @Description Replaces its members. Each is a member of the organization who is not removed.
// @Tags    Organizations
// @Produce json
// @Id      setOrgGroupMembers
// @Param   orgID path string true "organization ID"
// @Param   groupID path string true "group ID"
// @Param   body body orgdto.SetGroupMembersReq true "request data"
// @Success 200 {object} orgdto.GroupDataResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/groups/{groupID}/members [put]
func (h *Handler) SetGroupMembers(ctx *gin.Context) {
	auth, err := h.adminAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := orgdto.NewSetGroupMembersReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.orgUC.SetGroupMembers(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramGroupID),
		req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
