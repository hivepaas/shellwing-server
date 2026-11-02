package syncdto

import (
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
)

// SpaceResp is a space and its totals.
type SpaceResp struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Revision    int64  `json:"revision"`
	BytesUsed   int64  `json:"bytesUsed"`
	RecordCount int    `json:"recordCount"`
	HostCount   int    `json:"hostCount"`
}

func TransformSpace(space *entity.Space, totals *entity.SyncSpace) *SpaceResp {
	return &SpaceResp{
		ID:          space.ID,
		Type:        string(space.Type),
		Revision:    totals.Revision,
		BytesUsed:   totals.BytesUsed,
		RecordCount: totals.RecordCount,
		HostCount:   totals.HostCount,
	}
}

type ListSpacesResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*SpaceResp  `json:"data"`
}

// ObjectResp is one changed file. Deleted: it was removed, and the client drops its copy.
type ObjectResp struct {
	Name    string `json:"name"`
	ETag    string `json:"etag"`
	Size    int64  `json:"size"`
	Deleted bool   `json:"deleted"`
}

// ChangesResp is what happened after a revision.
type ChangesResp struct {
	Revision int64         `json:"revision"`
	More     bool          `json:"more"`
	Objects  []*ObjectResp `json:"objects"`
}

type ListChangesResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *ChangesResp  `json:"data"`
}

// PutObjectData is the file as it stands after a write.
type PutObjectData struct {
	ETag     string `json:"etag"`
	Revision int64  `json:"revision"`
}

type PutObjectResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *PutObjectData `json:"data"`
}

type DeleteObjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
