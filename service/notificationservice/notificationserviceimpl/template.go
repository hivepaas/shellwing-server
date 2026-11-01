package notificationserviceimpl

import (
	"context"
	"io"
	"sync"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/service/notificationservice"
)

type Template interface {
	Execute(wr io.Writer, data any) error
}

var (
	templateMap = map[notificationservice.TemplateType]map[notificationservice.TemplateName]Template{}
	mu          sync.Mutex
)

func (s *service) GetTemplate(
	ctx context.Context,
	db database.IDB,
	typ notificationservice.TemplateType,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	mu.Lock()
	defer mu.Unlock()

	mapTplByName, exists := templateMap[typ]
	if !exists {
		mapTplByName = make(map[notificationservice.TemplateName]Template, 5) //nolint:mnd
		templateMap[typ] = mapTplByName
	}

	tpl, exists = mapTplByName[name]
	if exists {
		return tpl, nil
	}

	switch typ {
	case notificationservice.TemplateTypeEmail:
		tpl, err = s.loadEmailTemplate(ctx, db, name)
	case notificationservice.TemplateTypeSlack:
		tpl, err = s.loadSlackTemplate(ctx, db, name)
	case notificationservice.TemplateTypeDiscord:
		tpl, err = s.loadDiscordTemplate(ctx, db, name)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The server has no notification templates yet. One asked for is refused,
	// not cached and handed back as nil for its caller's Execute to panic on.
	if tpl == nil {
		return nil, hperrors.NewNotFound("Template").WithExtraDetail("No %s template %q.", typ, name)
	}
	mapTplByName[name] = tpl

	return tpl, nil
}

func (s *service) loadEmailTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	return tpl, nil
}

func (s *service) loadSlackTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	return tpl, nil
}

func (s *service) loadDiscordTemplate(
	_ context.Context,
	_ database.IDB,
	name notificationservice.TemplateName,
) (tpl Template, err error) {
	return tpl, nil
}
