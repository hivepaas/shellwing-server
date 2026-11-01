package queueimpl

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

// taskTypeTimeouts caps how long a task of each type may run when it carries no timeout of its
// own. The cap matters beyond the task itself: a task holds its database transaction open for its
// whole run, and Postgres cannot reclaim any row version older than the longest open transaction.
// One task running for the old blanket three hours therefore bloats the entire database for three
// hours, not only the tasks table.
//
// The values are deliberately generous. Cutting a task short that used to finish is worse than
// the bloat it avoids, so types whose work is genuinely open-ended - running whatever job a user
// configured - keep the long ceiling. What this buys is a much lower ceiling for the types whose
// duration is predictable.
const (
	timeoutDummy = 5 * time.Minute
)

var taskTypeTimeouts = map[base.TaskType]time.Duration{
	// A test task that should never be slow.
	base.TaskTypeDummy:        timeoutDummy,
	base.TaskTypeSchedJobExec: taskDefaultTimeout,
}

// resolveTaskTimeout picks the timeout for a task: what the task asks for, else the ceiling for
// its type, else the blanket default.
func resolveTaskTimeout(task *entity.Task) time.Duration {
	if timeout := task.Config.Timeout.ToDuration(); timeout > 0 {
		return timeout
	}
	if timeout, ok := taskTypeTimeouts[task.Type]; ok && timeout > 0 {
		return timeout
	}
	return taskDefaultTimeout
}
