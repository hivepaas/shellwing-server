package systemhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/base"
	_ "github.com/hivepaas/shellwing-server/hperrors"
	_ "github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

// ListTask Lists tasks
// @Summary Lists tasks
// @Description Lists tasks
// @Tags    System tasks
// @Produce json
// @Id      listTask
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Param   fromDate query string false "`fromDate=YYYY-MM-DD`"
// @Param   scopeOnly query bool false "`scopeOnly=true` leaves out the tasks of what the scope holds"
// @Param   status query string false "`status=<status>`, comma separated"
// @Param   targetId query string false "`targetId=<ID>`, comma separated: the objects the tasks act on"
// @Param   toDate query string false "`toDate=YYYY-MM-DD`"
// @Param   type query string false "`type=<task type>`, comma separated"
// @Success 200 {object} taskdto.ListTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks [get]
func (h *Handler) ListTask(ctx *gin.Context) {
	h.taskHandler.ListTask(ctx, base.ObjectScopeGlobal)
}

// GetTask Gets task
// @Summary Gets task
// @Description Gets task
// @Tags    System tasks
// @Produce json
// @Id      getTask
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID} [get]
func (h *Handler) GetTask(ctx *gin.Context) {
	h.taskHandler.GetTask(ctx, base.ObjectScopeGlobal)
}

// GetTaskStatus Gets task status
// @Summary Gets task status
// @Description Gets task status
// @Tags    System tasks
// @Produce json
// @Id      getTaskStatus
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/status [get]
func (h *Handler) GetTaskStatus(ctx *gin.Context) {
	h.taskHandler.GetTaskStatus(ctx, base.ObjectScopeGlobal)
}

// GetTaskLogs Gets task logs
// @Summary Gets task logs
// @Description Gets task logs
// @Tags    System tasks
// @Produce json
// @Id      getTaskLogs
// @Param   itemID path string true "task ID"
// @Param   duration query string false "`duration=1h` for the logs within that period"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   tail query int false "`tail=1000` for the last 1000 lines"
// @Success 200 {object} taskdto.GetTaskLogsReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/logs [get]
func (h *Handler) GetTaskLogs(ctx *gin.Context) {
	h.taskHandler.GetTaskLogs(ctx, base.ObjectScopeGlobal)
}

// CancelTask Cancels task
// @Summary Cancels task
// @Description Cancels task
// @Tags    System tasks
// @Produce json
// @Id      cancelTask
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.CancelTaskReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/cancel [post]
func (h *Handler) CancelTask(ctx *gin.Context) {
	h.taskHandler.CancelTask(ctx, base.ObjectScopeGlobal)
}
