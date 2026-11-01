package useruc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) GetUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.GetUserReq,
) (*userdto.GetUserResp, error) {
	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}
	if req.GetAccesses {
		// The module accesses and capabilities are both read from these rows.
		loadOpts = append(loadOpts, bunex.SelectRelation("Accesses"))
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID, loadOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := userdto.TransformUserDetails(user)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &userdto.GetUserResp{
		Data: resp,
	}, nil
}
