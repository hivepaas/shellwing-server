package base

type SystemEventType string

const (
	// SystemEventSyncChanged says a space has a new revision, to every instance. Its payload is "<space id> <revision>".
	SystemEventSyncChanged SystemEventType = "sync-changed"
)
