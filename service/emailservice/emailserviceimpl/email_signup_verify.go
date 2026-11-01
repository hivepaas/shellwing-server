package emailserviceimpl

import (
	"context"
	"html/template"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bbpool"
	"github.com/hivepaas/shellwing-server/pkg/email"
	"github.com/hivepaas/shellwing-server/service/emailservice"
)

const signupVerifySubject = "Finish creating your shellwing account"

func (s *service) SendMailSignupVerify(
	ctx context.Context,
	db database.IDB,
	data *emailservice.EmailDataSignupVerify,
) error {
	template, err := s.GetTemplate(ctx, db, emailservice.TemplateNameSignupVerify)
	if err != nil {
		return hperrors.Wrap(err)
	}

	subject, content, err := renderSignupVerify(template, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = email.SendMail(ctx, data.Email, data.Recipients, subject, content)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// renderSignupVerify returns the subject and the HTML body of a sign-up link email. A subject the caller set wins.
func renderSignupVerify(
	tpl *template.Template,
	data *emailservice.EmailDataSignupVerify,
) (subject string, content string, err error) {
	if data == nil || data.SignupLink == "" {
		return "", "", hperrors.NewArgumentInvalid("sign-up link").
			WithMsgLog("a sign-up email needs its link")
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = tpl.Execute(buf, data)
	if err != nil {
		return "", "", hperrors.Wrap(err)
	}

	return gofn.Coalesce(data.Subject, signupVerifySubject), buf.String(), nil
}
