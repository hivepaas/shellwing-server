package taskdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/entity/cacheentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/copier"
)

type GetTaskReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-"`
}

func NewGetTaskReq() *GetTaskReq {
	return &GetTaskReq{}
}

func (req *GetTaskReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *TaskResp     `json:"data"`
}

type TaskResp struct {
	ID        string             `json:"id"`
	Type      base.TaskType      `json:"type"`
	Status    base.TaskStatus    `json:"status"`
	Config    entity.TaskConfig  `json:"config"` // NOTE: use entity's type directly here, may need refactor
	TargetJob *TaskTargetJobResp `json:"targetJob"`
	LastError string             `json:"lastError"`
	UpdateVer int                `json:"updateVer"`

	ScopeUser *basedto.UserBaseResp `json:"scopeUser,omitempty"`

	RunAt     *time.Time `json:"runAt" copy:",nilonzero"`
	RetryAt   *time.Time `json:"retryAt" copy:",nilonzero"`
	StartedAt *time.Time `json:"startedAt" copy:",nilonzero"`
	EndedAt   *time.Time `json:"endedAt" copy:",nilonzero"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskTargetJobResp struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TransformTask(
	task *entity.Task,
	taskInfo *cacheentity.TaskInfo,
	refObjects *entity.RefObjects,
) (resp *TaskResp, err error) {
	if err = copier.Copy(&resp, &task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if taskInfo != nil {
		resp.Status = taskInfo.Status
		if taskInfo.Status == base.TaskStatusInProgress {
			resp.StartedAt = &taskInfo.StartedAt
		}
	}

	if resp.Status == base.TaskStatusInProgress || resp.Status == base.TaskStatusFailed {
		resp.LastError = task.GetLastError()
	}

	TransformTaskScopeObject(task, refObjects, resp)

	return resp, nil
}

func TransformTaskScopeObject(
	task *entity.Task,
	refObjects *entity.RefObjects,
	resp *TaskResp,
) {
	if task.ObjectID == "" {
		return
	}

	var userID string
	switch task.Scope {
	case base.ObjectScopeUser:
		userID = task.ObjectID
	case base.ObjectScopeGlobal:
	}

	if userID != "" {
		refResp := basedto.TransformUserBase(refObjects.RefUsers[userID])
		if refResp == nil {
			refResp = basedto.NewMissingUser(userID)
		}
		resp.ScopeUser = refResp
	}
}
