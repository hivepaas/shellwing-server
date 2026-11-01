package taskuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/taskservice"
	"github.com/hivepaas/shellwing-server/usecase/taskuc/taskdto"
)

// spyTaskService keeps the task requests it was asked, and finds nothing.
type spyTaskService struct {
	taskservice.Service
	asked []*taskservice.GetTaskReq
}

func (s *spyTaskService) GetTask(_ context.Context, _ database.IDB, req *taskservice.GetTaskReq) (
	*taskservice.GetTaskResp, error,
) {
	s.asked = append(s.asked, req)
	return nil, hperrors.NewNotFound("Task")
}

// A task is read within the scope it was reached through, as its status, logs
// and cancel are: through one user's path, another user's task - or the
// system's - is not found, whatever its id.
func TestGetTaskLooksInTheScopeItWasReachedThrough(t *testing.T) {
	tasks := &spyTaskService{}
	uc := &UC{taskService: tasks}
	scope := entity.NewObjectScopeUser("u1")

	_, err := uc.GetTask(context.Background(), adminAuth(), &taskdto.GetTaskReq{Scope: scope, ID: "task-of-b"})

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	if assert.Len(t, tasks.asked, 1) {
		assert.Same(t, scope, tasks.asked[0].Scope)
		assert.Equal(t, "task-of-b", tasks.asked[0].ID)
	}
}
