package reslinkservice

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Service interface {
	SetLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string, options ...SetLinkOption) error
	AddLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string, options ...SetLinkOption) error
	RemoveLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string) error
}
