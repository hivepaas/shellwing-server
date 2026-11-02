package hperrors

// Errors for syncing files.
var (
	// ErrSyncObjectChanged refuses a conditional write: the file is not as the writer last saw it. It reads the file
	// again, merges, and writes once more.
	ErrSyncObjectChanged = NewErr(ErrPreconditionFailed, "ERR_SYNC_OBJECT_CHANGED")
	// ErrSyncObjectNameInvalid is a name outside the layout clients sync.
	ErrSyncObjectNameInvalid = NewErr(ErrArgumentInvalid, "ERR_SYNC_OBJECT_NAME_INVALID")
	// ErrSyncRecordInvalid is a record whose clear part cannot be read, or names another record.
	ErrSyncRecordInvalid = NewErr(ErrArgumentInvalid, "ERR_SYNC_RECORD_INVALID")
	// ErrSyncObjectTooLarge is a file above the server's limit for one.
	ErrSyncObjectTooLarge = NewErr(ErrArgumentInvalid, "ERR_SYNC_OBJECT_TOO_LARGE")
	// ErrSyncSpaceFull refuses a write that would take the space past its limit.
	ErrSyncSpaceFull = NewErr(ErrConflict, "ERR_SYNC_SPACE_FULL")
)
