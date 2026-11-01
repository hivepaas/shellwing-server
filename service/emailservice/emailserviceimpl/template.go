package emailserviceimpl

import (
	"context"
	"html/template"
	"sync"

	"github.com/hivepaas/shellwing-server/assets"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/emailservice"
)

var (
	templateMap = map[emailservice.TemplateName]*template.Template{}
	mu          sync.Mutex
)

func (s *service) GetTemplate(
	_ context.Context,
	_ database.IDB,
	name emailservice.TemplateName,
) (tpl *template.Template, err error) {
	mu.Lock()
	defer mu.Unlock()

	if tpl, exists := templateMap[name]; exists {
		return tpl, nil
	}

	switch name {
	case emailservice.TemplateNamePasswordReset:
		tpl, err = template.ParseFS(assets.GetTemplatesFS(), "email/password_reset.html")
	case emailservice.TemplateNameUserInvite:
		tpl, err = template.ParseFS(assets.GetTemplatesFS(), "email/user_invite.html")
	case emailservice.TemplateNameSignupVerify:
		tpl, err = template.ParseFS(assets.GetTemplatesFS(), "email/signup_verify.html")
	default:
		return nil, hperrors.NewArgumentInvalid("email template").
			WithMsgLog("unknown email template '%s'", name)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	templateMap[name] = tpl

	return tpl, nil
}
