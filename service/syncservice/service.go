package syncservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/entity"
)

// Condition is what a write requires of the file. The zero value requires nothing.
type Condition struct {
	// IfMatch: the file exists, with this ETag.
	IfMatch string
	// IfAbsent: the file does not exist.
	IfAbsent bool
}

// Changes are what happened after a revision.
type Changes struct {
	// Revision is where the client stands once it has these: pass it as since next time.
	Revision int64
	Objects  []*entity.SyncObject
	// More says there are further changes after Revision: ask again at once.
	More bool
}

type Service interface {
	// EnsurePersonalSpace returns the user's personal space, making it the first time.
	EnsurePersonalSpace(ctx context.Context, userID string) (*entity.Space, error)
	GetSpace(ctx context.Context, spaceID string) (*entity.Space, error)
	GetTotals(ctx context.Context, spaceID string) (*entity.SyncSpace, error)

	// Write stores a file, under the condition. It returns its line. Writing what the file already holds changes
	// nothing and does not raise the revision.
	Write(ctx context.Context, spaceID, name string, data []byte, cond Condition) (*entity.SyncObject, error)
	// Stat returns a live file's line, or not found.
	Stat(ctx context.Context, spaceID, name string) (*entity.SyncObject, error)
	// Read returns a live file's line and content, or not found.
	Read(ctx context.Context, spaceID, name string) (*entity.SyncObject, []byte, error)
	// Remove removes a file, under the condition. A file that is not there is no error, unless IfMatch names one.
	Remove(ctx context.Context, spaceID, name string, cond Condition) error
	// ChangesSince lists up to limit changes after the revision. Since 0 leaves out what was removed.
	ChangesSince(ctx context.Context, spaceID string, since int64, limit int) (*Changes, error)

	// Start and Stop run the cleaning of the store.
	Start(ctx context.Context) error
	Stop() error
}
