package synchandler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/interface/api/handler"
	"github.com/hivepaas/shellwing-server/interface/api/handler/authhandler"
	"github.com/hivepaas/shellwing-server/pkg/safego"
	"github.com/hivepaas/shellwing-server/service/syncservice"
	"github.com/hivepaas/shellwing-server/service/systemeventbusservice"
	"github.com/hivepaas/shellwing-server/usecase/syncuc"
	_ "github.com/hivepaas/shellwing-server/usecase/syncuc/syncdto"
)

const (
	headerETag        = "ETag"
	headerIfMatch     = "If-Match"
	headerIfNoneMatch = "If-None-Match"
	paramSpaceID      = "spaceID"

	eventsPingInterval = 30 * time.Second
	eventsWriteTimeout = 10 * time.Second
	// A connection does not outlive this: reconnecting checks the person's access again.
	eventsMaxAge = time.Hour
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	syncUC      *syncuc.UC
	eventBus    systemeventbusservice.Service
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	syncUC *syncuc.UC,
	eventBus systemeventbusservice.Service,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		syncUC:      syncUC,
		eventBus:    eventBus,
	}
}

func quoted(etag string) string {
	return `"` + etag + `"`
}

func unquoted(etag string) string {
	return strings.Trim(strings.TrimSpace(etag), `"`)
}

// conditionOf reads a write's condition from its headers. Only `If-None-Match: *` is a condition on a write.
func conditionOf(ctx *gin.Context) syncservice.Condition {
	return syncservice.Condition{
		IfMatch:  unquoted(ctx.GetHeader(headerIfMatch)),
		IfAbsent: strings.TrimSpace(ctx.GetHeader(headerIfNoneMatch)) == "*",
	}
}

// ListSyncSpaces lists the person's spaces
// @Summary Lists the person's spaces
// @Description The spaces the person syncs, with their revisions and totals. The personal one is made the first time
// @Description this is called.
// @Tags    Sync
// @Produce json
// @Id      listSyncSpaces
// @Success 200 {object} syncdto.ListSpacesResp
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces [get]
func (h *Handler) ListSyncSpaces(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.syncUC.ListSpaces(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ListSyncChanges lists the files changed after a revision
// @Summary Lists the files changed after a revision
// @Description What changed after `since`, oldest first. Pass the answer's `revision` as `since` next time; `more`
// @Description says to ask again at once. With `since` 0, removed files are left out.
// @Tags    Sync
// @Produce json
// @Id      listSyncChanges
// @Param   spaceID path string true "space ID"
// @Param   since query int false "the revision the client has seen"
// @Param   limit query int false "the most changes to answer with"
// @Success 200 {object} syncdto.ListChangesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces/{spaceID}/changes [get]
func (h *Handler) ListSyncChanges(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	since, err := parseQueryInt(ctx.Query("since"), "since")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	limit, err := parseQueryInt(ctx.Query("limit"), "limit")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.syncUC.ListChanges(h.RequestCtx(ctx), auth, ctx.Param(paramSpaceID), since, int(limit))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func parseQueryInt(raw, name string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, hperrors.NewArgumentInvalid(name)
	}
	return value, nil
}

// GetSyncObject reads a file
// @Summary Reads a file
// @Description The file's bytes, with its `ETag`. With `If-None-Match` naming that ETag: 304 and no body.
// @Tags    Sync
// @Produce application/octet-stream
// @Id      getSyncObject
// @Param   spaceID path string true "space ID"
// @Param   name query string true "the file's name, such as records/<id>.sw"
// @Param   If-None-Match header string false "the ETag the client holds"
// @Success 200
// @Success 304
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces/{spaceID}/object [get]
func (h *Handler) GetSyncObject(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	held := unquoted(ctx.GetHeader(headerIfNoneMatch))
	object, data, err := h.syncUC.GetObject(h.RequestCtx(ctx), auth, ctx.Param(paramSpaceID), ctx.Query("name"), held)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.Header(headerETag, quoted(object.ETag))
	if held != "" && object.ETag == held {
		ctx.Status(http.StatusNotModified)
		ctx.Writer.WriteHeaderNow()
		return
	}
	ctx.Data(http.StatusOK, "application/octet-stream", data)
}

// PutSyncObject writes a file
// @Summary Writes a file
// @Description Writes the body as the file. `If-Match` writes only over that ETag, and `If-None-Match: *` only if there
// @Description is no such file; refused, the answer is 412 with `ERR_SYNC_OBJECT_CHANGED`. Without either the write
// @Description goes through.
// @Tags    Sync
// @Accept  application/octet-stream
// @Produce json
// @Id      putSyncObject
// @Param   spaceID path string true "space ID"
// @Param   name query string true "the file's name, such as records/<id>.sw"
// @Param   If-Match header string false "write only over this ETag"
// @Param   If-None-Match header string false "* to write only a new file"
// @Param   body body string true "the file's bytes"
// @Success 200 {object} syncdto.PutObjectResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces/{spaceID}/object [put]
func (h *Handler) PutSyncObject(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	h.LimitRequestBody(ctx, config.Current().Sync.MaxObjectBytes)
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			h.RenderError(ctx, hperrors.Wrap(hperrors.ErrSyncObjectTooLarge))
			return
		}
		h.RenderError(ctx, hperrors.NewArgumentInvalid("body").WithCause(err))
		return
	}
	resp, err := h.syncUC.PutObject(h.RequestCtx(ctx), auth, ctx.Param(paramSpaceID), ctx.Query("name"), data,
		conditionOf(ctx))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.Header(headerETag, quoted(resp.Data.ETag))
	ctx.JSON(http.StatusOK, resp)
}

// DeleteSyncObject removes a file
// @Summary Removes a file
// @Description Removes the file. A file that is not there is no error, unless `If-Match` names a version of it.
// @Tags    Sync
// @Produce json
// @Id      deleteSyncObject
// @Param   spaceID path string true "space ID"
// @Param   name query string true "the file's name, such as records/<id>.sw"
// @Param   If-Match header string false "remove only this ETag"
// @Success 200 {object} syncdto.DeleteObjectResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces/{spaceID}/object [delete]
func (h *Handler) DeleteSyncObject(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.syncUC.DeleteObject(h.RequestCtx(ctx), auth, ctx.Param(paramSpaceID), ctx.Query("name"),
		conditionOf(ctx))
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// syncEvent is what the events stream sends: the space's revision is now at least this.
type syncEvent struct {
	Revision string `json:"revision"`
}

// StreamSyncEvents tells a client when a space changes
// @Summary Tells a client when a space changes
// @Description A WebSocket. Each message names a revision the space has reached: the client then asks for the changes.
// @Description A message can be lost, so a client also asks when it connects, and every so often.
// @Tags    Sync
// @Id      streamSyncEvents
// @Param   spaceID path string true "space ID"
// @Success 101
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /sync/spaces/{spaceID}/events [get]
func (h *Handler) StreamSyncEvents(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	spaceID := ctx.Param(paramSpaceID)
	if err = h.syncUC.CheckAccess(h.RequestCtx(ctx), auth, spaceID); err != nil {
		h.RenderError(ctx, err)
		return
	}
	conn, err := h.UpgradeWebsocket(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer func() { _ = conn.Close() }()

	events, unsubscribe := h.eventBus.Subscribe(base.SystemEventSyncChanged)
	defer unsubscribe()

	// Reading is what notices the client going away. It sends nothing else that matters.
	closed := make(chan struct{})
	safego.Go("syncEventsRead", func() {
		defer close(closed)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})

	ping := time.NewTicker(eventsPingInterval)
	defer ping.Stop()
	maxAge := time.NewTimer(eventsMaxAge)
	defer maxAge.Stop()
	prefix := spaceID + " "
	for {
		select {
		case <-closed:
			return
		case <-maxAge.C:
			return
		case <-ping.C:
			err = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(eventsWriteTimeout))
			if err != nil {
				return
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			revision, found := strings.CutPrefix(event.Payload, prefix)
			if !found {
				continue
			}
			_ = conn.SetWriteDeadline(time.Now().Add(eventsWriteTimeout))
			if err = conn.WriteJSON(&syncEvent{Revision: revision}); err != nil {
				return
			}
		}
	}
}
