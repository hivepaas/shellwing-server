package queueimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/tasks/queue"
)

// What runs after a task's transaction runs past the recover around its
// execution: a panic there is contained, not let take the process down.
func TestAPanicAfterTheTransactionIsContained(t *testing.T) {
	q := &taskQueue{logger: logging.GlobalLogger()}
	data := &queue.TaskExecData{Task: &entity.Task{Type: base.TaskTypeSchedJobExec}}
	data.OnPostTx(func() { panic("boom") })

	assert.NotPanics(t, func() { q.runPostTx(data) })
}

func TestWhatRunsAfterTheTransactionRuns(t *testing.T) {
	q := &taskQueue{logger: logging.GlobalLogger()}
	ran := false
	data := &queue.TaskExecData{Task: &entity.Task{Type: base.TaskTypeSchedJobExec}}
	data.OnPostTx(func() { ran = true })

	q.runPostTx(data)
	q.runPostTx(nil) // a task that was not run has nothing to run after

	assert.True(t, ran)
}
