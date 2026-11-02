package entity

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/base"
)

type Space struct {
	ID          string         `bun:",pk" json:"id"`
	Type        base.SpaceType `json:"type"`
	OrgID       string         `bun:",nullzero" json:"orgId,omitempty"`
	OwnerUserID string         `bun:",nullzero" json:"ownerUserId,omitempty"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`
}

// GetID implements IDEntity interface
func (s *Space) GetID() string {
	return s.ID
}

// SyncSpace is a space's totals. They count what is live: a tombstone or a removed file is not in them, but for
// BytesUsed, which counts what the store holds for the live lines, tombstones included.
type SyncSpace struct {
	SpaceID     string `bun:",pk" json:"spaceId"`
	Revision    int64  `json:"revision"`
	BytesUsed   int64  `json:"bytesUsed"`
	RecordCount int    `json:"recordCount"`
	HostCount   int    `json:"hostCount"`

	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

// SyncObject is the line of one synced file. Its content is in the store, at SyncBlobKey.
type SyncObject struct {
	SpaceID   string `bun:",pk" json:"spaceId"`
	Name      string `bun:",pk" json:"name"`
	ETag      string `bun:"etag" json:"etag"`
	Size      int64  `json:"size"`
	Kind      string `bun:",nullzero" json:"kind,omitempty"`
	Tombstone bool   `json:"tombstone"`
	Revision  int64  `json:"revision"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	// DeletedAt is not bun's soft delete: the changes since a revision have to see these lines.
	DeletedAt time.Time `bun:",nullzero" json:"deletedAt,omitzero"`
}

// IsLive tells a file that exists from one removed.
func (o *SyncObject) IsLive() bool {
	return o != nil && o.DeletedAt.IsZero()
}

// SyncTrash is content the store may drop once DeleteAfter has passed, if no line names it by then.
type SyncTrash struct {
	bun.BaseModel `bun:"table:sync_trash,alias:sync_trash"`

	SpaceID     string    `bun:",pk"`
	Name        string    `bun:",pk"`
	ETag        string    `bun:"etag,pk"`
	DeleteAfter time.Time ``
}

// SyncBlobKey is where the store keeps one version of a file. Content never changes under a key.
func SyncBlobKey(spaceID, name, etag string) string {
	return "spaces/" + spaceID + "/" + name + "/" + etag
}
