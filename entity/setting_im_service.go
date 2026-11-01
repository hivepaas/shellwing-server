package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	CurrentIMServiceVersion = 1
)

var _ = registerSettingParser(base.SettingTypeIMService, &imServiceParser{})

type imServiceParser struct {
}

func (s *imServiceParser) New() SettingData {
	return &IMService{}
}

type IMService struct {
	Slack   *IMSlack   `json:"slack,omitempty"`
	Discord *IMDiscord `json:"discord,omitempty"`
}

type IMSlack struct {
	Webhook EncryptedField `json:"webhook"`
}

type IMDiscord struct {
	Webhook EncryptedField `json:"webhook"`
}

func (s *IMService) GetType() base.SettingType {
	return base.SettingTypeIMService
}

func (s *IMService) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *IMService) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *IMService) Decrypt() error {
	if s.Slack != nil {
		_, err := s.Slack.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	if s.Discord != nil {
		_, err := s.Discord.Webhook.GetPlain()
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Setting) AsIMService() (*IMService, error) {
	return parseSettingAs[*IMService](s)
}

func (s *Setting) MustAsIMService() *IMService {
	return gofn.Must(s.AsIMService())
}
