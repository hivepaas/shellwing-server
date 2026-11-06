package syncdto

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
)

// SpaceResp is a space and its totals. A collection also has its organization, name, key version and what the person
// may do in it.
type SpaceResp struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Revision    int64  `json:"revision"`
	BytesUsed   int64  `json:"bytesUsed"`
	RecordCount int    `json:"recordCount"`
	HostCount   int    `json:"hostCount"`

	OrgID      string                    `json:"orgId,omitempty"`
	Name       string                    `json:"name,omitempty"`
	KeyVersion int                       `json:"keyVersion,omitempty"`
	Permission base.CollectionPermission `json:"permission,omitempty"`
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

func TransformCollection(collection *entity.Collection, permission base.CollectionPermission,
	totals *entity.SyncSpace) *SpaceResp {
	resp := TransformSpace(&entity.Space{ID: collection.ID, Type: base.SpaceTypeCollection}, totals)
	resp.OrgID = collection.OrgID
	resp.Name = collection.Name
	resp.KeyVersion = collection.KeyVersion
	resp.Permission = permission
	return resp
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
