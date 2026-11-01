package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
)

// newStoredSetting builds a setting and returns it the way a row read back from
// the database looks, with no parsed cache.
func newStoredSetting(t *testing.T, data SettingData) *Setting {
	t.Helper()

	useDataKey(t)
	setting := &Setting{ID: "set_1", Type: data.GetType()}
	assert.NoError(t, setting.SetData(data))
	return &Setting{ID: setting.ID, Type: setting.Type, Data: setting.Data}
}

func TestReencryptDataRewritesNestedSecrets(t *testing.T) {
	setting := newStoredSetting(t, &Email{
		SMTP: &EmailSMTP{Host: "smtp.example.com", Password: NewEncryptedField("smtp-password")},
		HTTP: &EmailHTTP{Endpoint: "https://mail.example.com", Password: NewEncryptedField("http-password")},
	})
	before := setting.Data

	changed, err := setting.ReencryptData()
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.NotEqual(t, before, setting.Data, "the ciphertext must actually change")

	reloaded := &Setting{Type: setting.Type, Data: setting.Data}
	parsed, err := reloaded.Parse()
	assert.NoError(t, err)

	email, _ := parsed.(*Email)
	smtpPassword, err := email.SMTP.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "smtp-password", smtpPassword)
	httpPassword, err := email.HTTP.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "http-password", httpPassword)
}

// A nil provider must not be walked into, and a setting with nothing encrypted
// must not be rewritten at all.
func TestReencryptDataSkipsWhatItShould(t *testing.T) {
	t.Run("no encrypted field means no change", func(t *testing.T) {
		// S3 is the only provider and it is unset, so there is nothing to walk.
		setting := newStoredSetting(t, &CloudStorage{})

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("empty data means no change", func(t *testing.T) {
		useDataKey(t)
		setting := &Setting{ID: "set_1", Type: base.SettingTypeCloudStorage}

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("an unset secret is left alone", func(t *testing.T) {
		setting := newStoredSetting(t, &Email{
			SMTP: &EmailSMTP{Host: "smtp.example.com", Password: NewEncryptedField("smtp-password")},
			// HTTP password deliberately unset
			HTTP: &EmailHTTP{Endpoint: "https://mail.example.com"},
		})

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.True(t, changed)
		assert.NotContains(t, setting.Data, "smtp-password")

		reloaded := &Setting{Type: setting.Type, Data: setting.Data}
		email := reloaded.MustAsEmail()
		assert.True(t, email.HTTP.Password.IsEmpty(), "an unset secret must stay unset")
	})
}
