package collectionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/basedto"
	_ "github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc/collectiondto"
)

const (
	paramOrgID        = "orgID"
	paramCollectionID = "collectionID"
)

type Handler struct {
	*handler.BaseHandler
	authHandler  *authhandler.Handler
	collectionUC *collectionuc.UC
}

func New(baseHandler *handler.BaseHandler, authHandler *authhandler.Handler,
	collectionUC *collectionuc.UC) *Handler {
	return &Handler{BaseHandler: baseHandler, authHandler: authHandler, collectionUC: collectionUC}
}

// auth is a person signed in to the web app, or their app: both manage collections, as the person's role and
// permissions allow, which is the use case's to check.
func (h *Handler) auth(ctx *gin.Context) (*basedto.Auth, error) {
	return h.authHandler.GetSyncAuth(ctx) //nolint:wrapcheck // the handler's own
}

// render answers with the use case's result, or its error.
func (h *Handler) render(ctx *gin.Context, status int, resp any, err error) {
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(status, resp)
}

// ListCollections lists the organization's collections
// @Summary Lists the organization's collections
// @Description Those the person may use, with what they may do and whether they hold the current key; to an admin,
// @Description every one. Those who manage a collection's access also see how many wait for its key, and whether it
// @Description needs a new one.
// @Tags    Collections
// @Produce json
// @Id      listCollections
// @Param   orgID path string true "organization ID"
// @Success 200 {object} collectiondto.ListCollectionsResp
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections [get]
func (h *Handler) ListCollections(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.List(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID))
	h.render(ctx, http.StatusOK, resp, err)
}

// CreateCollection makes a collection
// @Summary Makes a collection in the organization
// @Description For owners and admins. It has no key until an app with Manage makes one; an admin who makes it is
// @Description granted Manage. Its name is unique in the organization, whatever its case.
// @Tags    Collections
// @Produce json
// @Id      createCollection
// @Param   orgID path string true "organization ID"
// @Param   body body collectiondto.CollectionReq true "request data"
// @Success 201 {object} collectiondto.CollectionDataResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections [post]
func (h *Handler) CreateCollection(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := collectiondto.NewCollectionReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.Create(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), req)
	h.render(ctx, http.StatusCreated, resp, err)
}

// GetCollection returns a collection
// @Summary Returns a collection
// @Description To those who manage its access, also its grants and everybody with access: what they may do, whether
// @Description they hold the key, and their identity, whose fingerprint is compared before giving it to them.
// @Tags    Collections
// @Produce json
// @Id      getCollection
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Success 200 {object} collectiondto.GetCollectionResp
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID} [get]
func (h *Handler) GetCollection(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.Get(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramCollectionID))
	h.render(ctx, http.StatusOK, resp, err)
}

// UpdateCollection changes a collection
// @Summary Changes a collection's name and description
// @Description For those who manage its access.
// @Tags    Collections
// @Produce json
// @Id      updateCollection
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Param   body body collectiondto.CollectionReq true "request data"
// @Success 200 {object} collectiondto.CollectionDataResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID} [put]
func (h *Handler) UpdateCollection(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := collectiondto.NewCollectionReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.Update(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramCollectionID),
		req)
	h.render(ctx, http.StatusOK, resp, err)
}

// DeleteCollection deletes a collection
// @Summary Deletes a collection
// @Description For owners and admins. Nobody syncs it from then on; its files go after the store's trash delay.
// @Tags    Collections
// @Produce json
// @Id      deleteCollection
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Success 200 {object} collectiondto.DeleteCollectionResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID} [delete]
func (h *Handler) DeleteCollection(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.Delete(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramCollectionID))
	h.render(ctx, http.StatusOK, resp, err)
}

// SetCollectionGrants sets who may do what in a collection
// @Summary Sets a collection's grants
// @Description These grants, and no others: each to a member who is not removed or to a group, Connect, Edit or
// @Description Manage. For those who manage its access. People newly granted wait for the key until an app holding
// @Description it gives it to them; people who lose access make the collection need a new key.
// @Tags    Collections
// @Produce json
// @Id      setCollectionGrants
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Param   body body collectiondto.SetGrantsReq true "request data"
// @Success 200 {object} collectiondto.SetGrantsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID}/grants [put]
func (h *Handler) SetCollectionGrants(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := collectiondto.NewSetGrantsReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.SetGrants(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID),
		ctx.Param(paramCollectionID), req)
	h.render(ctx, http.StatusOK, resp, err)
}

// GetCollectionGrantable returns whom a collection can be granted to
// @Summary Returns whom a collection can be granted to
// @Description For those who manage its access: the organization's people who are not removed, and its groups, with
// @Description only what choosing them needs. The app's token reaches it.
// @Tags    Collections
// @Produce json
// @Id      getCollectionGrantable
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Success 200 {object} collectiondto.GetGrantableResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID}/grantable [get]
func (h *Handler) GetCollectionGrantable(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.Grantable(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID), ctx.Param(paramCollectionID))
	h.render(ctx, http.StatusOK, resp, err)
}

// GetCollectionKeys returns the person's wraps of a collection's keys
// @Summary Returns the person's wraps of a collection's keys
// @Description The current key version, and the person's wraps of each version, newest first, with who wrapped
// @Description each: their app checks the signature and opens it. Each fetch is in the audit log.
// @Tags    Collections
// @Produce json
// @Id      getCollectionKeys
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Success 200 {object} collectiondto.GetKeysResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID}/keys [get]
func (h *Handler) GetCollectionKeys(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.GetKeys(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID),
		ctx.Param(paramCollectionID))
	h.render(ctx, http.StatusOK, resp, err)
}

// CreateCollectionKey gives a collection a new key
// @Summary Gives a collection a new key
// @Description Version current + 1, made by an app with Manage: the first, or one replacing the current, which that
// @Description app holds. Wrapped for the caller and every owner with an identity (ERR_COLLECTION_WRAPS_MISSING),
// @Description and for anybody else with access the app chooses; the others wait for it. Another app making one
// @Description first gives ERR_COLLECTION_KEY_VERSION.
// @Tags    Collections
// @Produce json
// @Id      createCollectionKey
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Param   body body collectiondto.WrapsReq true "request data"
// @Success 201 {object} collectiondto.WrapsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID}/keys [post]
func (h *Handler) CreateCollectionKey(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := collectiondto.NewWrapsReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.CreateKey(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID),
		ctx.Param(paramCollectionID), req)
	h.render(ctx, http.StatusCreated, resp, err)
}

// AddCollectionWraps gives the current key to people waiting for it
// @Summary Gives a collection's current key to people waiting for it
// @Description Wrapped by the app of somebody who manages the collection's access and holds the key, once they
// @Description compared each person's fingerprint. Each wrap is checked: signed by the caller's identity, for the
// @Description person's, who has access and does not hold the key yet (ERR_COLLECTION_WRAP_INVALID).
// @Tags    Collections
// @Produce json
// @Id      addCollectionWraps
// @Param   orgID path string true "organization ID"
// @Param   collectionID path string true "collection ID"
// @Param   body body collectiondto.WrapsReq true "request data"
// @Success 201 {object} collectiondto.WrapsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /orgs/{orgID}/collections/{collectionID}/wraps [post]
func (h *Handler) AddCollectionWraps(ctx *gin.Context) {
	auth, err := h.auth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := collectiondto.NewWrapsReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.collectionUC.AddWraps(h.RequestCtx(ctx), auth, ctx.Param(paramOrgID),
		ctx.Param(paramCollectionID), req)
	h.render(ctx, http.StatusCreated, resp, err)
}
