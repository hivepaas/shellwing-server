package entity_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

// ArgsAsX and OutputAsX read like getters and are not: the first caller parses
// and caches into the task. Steps of a task can run side by side, so two of them
// reaching a cold cache at the same time has to be safe.
//
// Run under -race, which scripts/test.sh does; without the lock this fails there
// rather than here.
func TestTaskArgsAndOutputSurviveConcurrentReaders(t *testing.T) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeDummy}
	task.MustSetArgs(&entity.TaskDummyArgs{Sleep: timeutil.Duration(time.Second)})
	task.MustSetOutput(&entity.TaskDummyOutput{})

	// A task loaded from the database has the strings and no parsed cache, which
	// is the state where the first reader writes.
	cold := &entity.Task{ID: task.ID, Type: task.Type, Args: task.Args, Output: task.Output}

	const readers = 8
	var wg sync.WaitGroup
	results := make([]time.Duration, readers)

	for i := range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			args, err := cold.ArgsAsDummy()
			if err == nil && args != nil {
				results[i] = args.Sleep.ToDuration()
			}
			_, _ = cold.OutputAsDummy()
		}()
	}
	wg.Wait()

	for i, got := range results {
		assert.Equal(t, time.Second, got, "reader %d", i)
	}
}
