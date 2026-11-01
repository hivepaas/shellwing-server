package sessionuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/usecase/sessionuc/sessiondto"
)

// GetSignupOptions says whether a person may create their own account, so that the sign-in page offers it only
// when the request would be taken. It is a switch of the config, public like the login options beside it.
func (uc *UC) GetSignupOptions(
	_ context.Context,
	_ *sessiondto.GetSignupOptionsReq,
) (*sessiondto.GetSignupOptionsResp, error) {
	return &sessiondto.GetSignupOptionsResp{
		Data: &sessiondto.SignupOptionsDataResp{
			SelfSignupEnabled: config.Current().Users.SelfSignupEnabled,
		},
	}, nil
}
