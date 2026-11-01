package auditlogdto

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
)

type ListAuditLogTypeReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewListAuditLogTypeReq() *ListAuditLogTypeReq {
	return &ListAuditLogTypeReq{}
}

type ListAuditLogTypeResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []base.AuditLogType `json:"data"`
}
