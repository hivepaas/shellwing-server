package emailuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/email"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
)

func (uc *UC) TestSendMail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.TestSendMailReq,
) (_ *emaildto.TestSendMailResp, err error) {
	conf := req.ToEntity()
	err = email.SendMail(ctx, conf, []string{req.TestRecipient}, req.TestSubject, req.TestContent)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &emaildto.TestSendMailResp{}, nil
}
