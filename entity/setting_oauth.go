package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	CurrentOAuthVersion = 1
)

var _ = registerSettingParser(base.SettingTypeOAuth, &oauthParser{})

type oauthParser struct {
}

func (s *oauthParser) New() SettingData {
	return &OAuth{}
}

type OAuth struct {
	ClientID         string         `json:"clientId"`
	ClientSecret     EncryptedField `json:"clientSecret"`
	Organization     string         `json:"org,omitempty"`
	AuthURL          string         `json:"authURL,omitempty"`
	TokenURL         string         `json:"tokenURL,omitempty"`
	ProfileURL       string         `json:"profileURL,omitempty"`
	AutoDiscoveryURL string         `json:"autoDiscoveryURL,omitempty"`
	Scopes           []string       `json:"scopes,omitempty"`
	// TrustEmail has a sign-in with an OpenID Connect provider take the email it gives as the person's own, whether
	// or not it says email_verified: for an identity provider whose users and their emails the administrator manages.
	TrustEmail bool `json:"trustEmail,omitempty"`
}

func (s *OAuth) GetType() base.SettingType {
	return base.SettingTypeOAuth
}

func (s *OAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *OAuth) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *OAuth) Decrypt() error {
	_, err := s.ClientSecret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsOAuth() (*OAuth, error) {
	return parseSettingAs[*OAuth](s)
}

func (s *Setting) MustAsOAuth() *OAuth {
	return gofn.Must(s.AsOAuth())
}
