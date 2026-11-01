package auditloguc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/usecase/auditloguc/auditlogdto"
)

func (uc *UC) ListAuditLogType(
	ctx context.Context,
	auth *basedto.Auth,
	req *auditlogdto.ListAuditLogTypeReq,
) (*auditlogdto.ListAuditLogTypeResp, error) {
	return &auditlogdto.ListAuditLogTypeResp{
		Data: base.AllAuditLogTypes,
	}, nil
}
