package sessionuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

func (uc *UC) GetMe(
	ctx context.Context,
	user *basedto.User,
	req *sessiondto.GetMeReq,
) (*sessiondto.GetMeResp, error) {
	loadOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	}
	if req.GetAccesses {
		// Must match useruc.GetUser: the module accesses and capabilities are
		// both read from these rows.
		loadOpts = append(loadOpts, bunex.SelectRelation("Accesses"))
	}

	dbUser, err := uc.userRepo.GetByID(ctx, uc.db, user.ID, loadOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	userResp, err := sessiondto.TransformUserDetails(dbUser)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData := &sessiondto.GetMeDataResp{
		User:     userResp,
		Timezone: timeutil.Location().String(),
	}

	if config.CurrentSystemInfo().NextStep != "" && user.IsAdmin() {
		sysStatus, err := uc.systemStatusRepo.Get(ctx, uc.db)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		config.SetInstallationStep(sysStatus.NextStep)
		respData.NextStep = string(sysStatus.NextStep)
	}

	if user.Status == base.UserStatusPending && user.TotpSecret == "" {
		respData.NextStep = nextStepMfaSetup
	}

	return &sessiondto.GetMeResp{
		Data: respData,
	}, nil
}
