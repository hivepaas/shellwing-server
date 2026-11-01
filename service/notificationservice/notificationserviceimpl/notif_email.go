package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bbpool"
	"github.com/hivepaas/shellwing-server/pkg/email"
	"github.com/hivepaas/shellwing-server/service/notificationservice"
)

const (
	emailSendRetryMax   = 2
	emailSendRetryDelay = 3 * time.Second
)

func (s *service) emailSendMsg(
	ctx context.Context,
	db database.IDB,
	emailSetting *entity.Setting,
	recipients []string,
	subject string,
	templateName notificationservice.TemplateName,
	templateData notificationservice.TemplateData,
) (err error) {
	if len(recipients) == 0 {
		return nil
	}
	if emailSetting == nil {
		return hperrors.NewMissing("Sender email account")
	}
	emailAcc := emailSetting.MustAsEmail()
	if emailAcc == nil {
		return hperrors.NewMissing("Sender email account")
	}

	template, err := s.GetTemplate(ctx, db, notificationservice.TemplateTypeEmail, templateName)
	if err != nil {
		return hperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, templateData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = email.SendMailRetry(ctx, emailAcc, recipients, subject, buf.String(),
		emailSendRetryMax, emailSendRetryDelay)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
