package entity

import (
	"time"

	"github.com/hivepaas/shellwing-server/base"
)

// AuditLog is one recorded action, kept so the question "who saw this, and when"
// has an answer.
//
// Rows are append-only. Nothing updates one, and the only deletion is the
// retention sweep - see sysCleanupDBOldAuditLogs, which will not let the entries
// that matter be aged out early.
//
// Several fields duplicate data that also lives in another table on purpose. They
// are a snapshot of the moment: the user may be deleted and the setting renamed
// long before anyone reads the entry, and an entry that has become a pair of
// meaningless identifiers is no use exactly when it is needed.
type AuditLog struct {
	ID       string               `bun:",pk" json:"id"`
	Scope    base.ObjectScopeType `json:"scope"`
	ObjectID string               `bun:",nullzero" json:"objectId,omitempty"`
	Type     base.AuditLogType    `json:"type"`
	Source   base.AuditLogSource  `json:"source,omitempty"`
	// Section is which part of the object was written - the settings tab, the
	// endpoint's slice of a project. It is a column rather than a field inside
	// Detail so it can be filtered and indexed, and because the listing drops
	// Detail once it grows past a hundred characters, which took the section with
	// it exactly on the entries that had most to say.
	Section string              `bun:",nullzero" json:"section,omitempty"`
	Result  base.AuditLogResult `json:"result"`

	ActorID   string           `json:"actorId"`
	ActorType base.SubjectType `json:"actorType"`
	ActorName string           `json:"actorName,omitempty"`
	// SessionUID identifies the session the action came from, so one compromised
	// session can be told apart from the owner's own use.
	SessionUID string `json:"sessionUid,omitempty"`

	ResType base.ResourceType `json:"resType,omitempty"`
	ResID   string            `json:"resId,omitempty"`
	ResName string            `json:"resName,omitempty"`

	// ClientIP is only as trustworthy as http_server.trusted_proxies; RemoteAddr
	// is the peer address, which a caller cannot choose. See pkg/reqinfo.
	ClientIP   string `json:"clientIp,omitempty"`
	RemoteAddr string `json:"remoteAddr,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	RequestID  string `json:"requestId,omitempty"`

	// Detail is free-form JSON for whatever a particular type needs, so a new
	// type does not need a new column. It must never hold a secret value, not
	// even shortened or hashed.
	Detail string `json:"detail,omitempty"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// GetID implements IDEntity interface
func (e *AuditLog) GetID() string {
	return e.ID
}

func (e *AuditLog) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if e.ObjectID != "" {
		switch e.Scope {
		case base.ObjectScopeUser:
			refIDs.RefUserIDs = append(refIDs.RefUserIDs, e.ObjectID)
		case base.ObjectScopeGlobal:
		}
	}
	if e.ActorID != "" && e.ActorType == base.SubjectTypeUser {
		refIDs.RefUserIDs = append(refIDs.RefUserIDs, e.ActorID)
	}
	return refIDs
}
