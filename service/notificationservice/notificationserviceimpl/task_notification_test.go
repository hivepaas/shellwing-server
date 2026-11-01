package notificationserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/notificationservice"
)

// countingSettings has no settings, and counts how often they are asked for.
type countingSettings struct {
	repository.SettingRepo
	lists int
}

func (r *countingSettings) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	r.lists++
	return nil, nil, nil
}

// A result that is not told - the same as the last, told within the
// notification's minimum interval - reads nothing: the sources it would be
// sent through used to be loaded first, a query for every result.
func TestAResultNotToldLoadsNoSources(t *testing.T) {
	settings := &countingSettings{}
	svc := &service{settingRepo: settings}
	data := &notificationservice.TaskResultNotificationReq{
		ActionSucceeded: true,
		Scope:           entity.NewObjectScopeGlobal(),
		RefObjects:      entity.NewRefObjects(),
		Notification: &entity.Notification{
			MinSendInterval: timeutil.Duration(10 * time.Minute),
			ViaEmail:        &entity.NotificationViaEmail{Enabled: true, UseDefault: true},
		},
		LastEvent:  "success",
		LastSendTs: time.Now().Add(-time.Minute),
	}

	resp, err := svc.NotifyForTaskResult(context.Background(), nil, data)
	assert.NoError(t, err)
	assert.False(t, resp.HasSend())
	assert.Zero(t, settings.lists)

	data.LastSendTs = time.Now().Add(-time.Hour)
	_, err = svc.NotifyForTaskResult(context.Background(), nil, data)
	assert.NoError(t, err)
	assert.Equal(t, 1, settings.lists, "one to be told loads them")
}
