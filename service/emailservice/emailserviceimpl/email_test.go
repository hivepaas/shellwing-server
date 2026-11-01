package emailserviceimpl

import (
	"context"
	"html"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/emailservice"
)

func testTemplate(t *testing.T, name emailservice.TemplateName) *template.Template {
	t.Helper()
	tpl, err := (&service{}).GetTemplate(context.Background(), nil, name)
	require.NoError(t, err)
	return tpl
}

// Every template the service names parses from the embedded files: a name added without its file, or a
// file with a syntax error, fails here rather than when the first mail of its kind goes out.
func TestEveryTemplateParses(t *testing.T) {
	names := []emailservice.TemplateName{
		emailservice.TemplateNamePasswordReset,
		emailservice.TemplateNameUserInvite,
		emailservice.TemplateNameSignupVerify,
	}
	s := &service{}
	for _, name := range names {
		tpl, err := s.GetTemplate(context.Background(), nil, name)
		assert.NoError(t, err, name)
		assert.NotNil(t, tpl, name)
	}

	_, err := s.GetTemplate(context.Background(), nil, "no-such-template")
	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid)
}

func TestSignupVerifyEmail(t *testing.T) {
	tpl := testTemplate(t, emailservice.TemplateNameSignupVerify)

	link := "https://shellwing.example.com/auth/sign-up?token=abc.def-ghi&next=%2Faccount%2Ftrial%3Fx%3D1"
	subject, content, err := renderSignupVerify(tpl, &emailservice.EmailDataSignupVerify{
		BaseTemplateData: emailservice.BaseTemplateData{Recipients: []string{"new@example.com"}},
		SignupLink:       link,
	})
	require.NoError(t, err)
	assert.Equal(t, "Finish creating your shellwing account", subject)
	// html/template writes '&' in an attribute as "&amp;", which the mail client reads back as '&'.
	assert.Contains(t, content, `href="`+html.EscapeString(link)+`"`)
	assert.Equal(t, 2, strings.Count(content, html.EscapeString(link)), "the button and the fallback link")
	assert.NotContains(t, content, "<no value>")

	_, _, err = renderSignupVerify(tpl, &emailservice.EmailDataSignupVerify{})
	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid)
}
