package initializer

import (
	"github.com/hivepaas/shellwing-server/tasks/taskdummy"
)

type WorkerInitializer struct {
}

// NOTE: these injections are required to make the task executors be available
func NewWorkerInitializer(
	_ *taskdummy.Executor,
) *WorkerInitializer {
	return nil
}
