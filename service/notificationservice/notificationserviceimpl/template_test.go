package notificationserviceimpl

import (
	"bytes"
	htmltemplate "html/template"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/assets"
)

func TestEmailPasswordResetTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/password_reset.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := struct {
		ResetPasswordLink string
	}{
		ResetPasswordLink: "https://hivepaas.io/reset?token=xyz",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Reset your password")
	assert.Contains(t, out, "badge-security")
	assert.Contains(t, out, "https://hivepaas.io/reset?token=xyz")
}

func TestEmailUserInviteTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/user_invite.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := struct {
		InviterName    string
		UserSignupLink string
	}{
		InviterName:    "Alice Smith",
		UserSignupLink: "https://hivepaas.io/invite?code=abc",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Join shellwing")
	assert.Contains(t, out, "badge-invite")
	assert.Contains(t, out, "Alice Smith")
	assert.Contains(t, out, "https://hivepaas.io/invite?code=abc")
}
