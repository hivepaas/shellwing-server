package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bbpool"
	"github.com/hivepaas/shellwing-server/pkg/im/imapi"
	"github.com/hivepaas/shellwing-server/service/notificationservice"
)

const (
	imSendRetryMax   = 2
	imSendRetryDelay = 2 * time.Second
)

func (s *service) imSendMsg(
	ctx context.Context,
	db database.IDB,
	imSetting *entity.Setting,
	templateType notificationservice.TemplateType,
	templateName notificationservice.TemplateName,
	templateData any,
) error {
	if imSetting == nil {
		return hperrors.NewMissing("IM service setting")
	}

	template, err := s.GetTemplate(ctx, db, templateType, templateName)
	if err != nil {
		return hperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, templateData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = imapi.SendMessageWithRetry(ctx, imSetting, buf.String(), imSendRetryMax, imSendRetryDelay)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
