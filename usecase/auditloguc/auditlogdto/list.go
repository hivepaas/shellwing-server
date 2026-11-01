package auditlogdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

type ListAuditLogReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`

	ScopeOnly bool `json:"-" mapstructure:"scopeOnly"`

	Type   []base.AuditLogType   `json:"-" mapstructure:"type"`
	Source []base.AuditLogSource `json:"-" mapstructure:"source"`
	// Section is filtered alongside Type rather than on its own: a section name
	// belongs to a type, and the index that serves this leads with type for that
	// reason.
	Section    []string              `json:"-" mapstructure:"section"`
	Result     []base.AuditLogResult `json:"-" mapstructure:"result"`
	ActorID    []string              `json:"-" mapstructure:"actorId"`
	ResourceID []string              `json:"-" mapstructure:"resourceId"`
	FromDate   timeutil.Date         `json:"-" mapstructure:"fromDate"`
	ToDate     timeutil.Date         `json:"-" mapstructure:"toDate"`
	Search     string                `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListAuditLogReq() *ListAuditLogReq {
	return &ListAuditLogReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListAuditLogReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAuditLogResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*AuditLogResp   `json:"data"`
}

func TransformAuditLogs(
	auditLogs []*entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp []*AuditLogResp, err error) {
	resp = make([]*AuditLogResp, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		item, err := TransformAuditLog(auditLog, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
