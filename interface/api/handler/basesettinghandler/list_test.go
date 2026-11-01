package basesettinghandler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/hivepaas/shellwing-server/usecase/settings/emailuc/emaildto"
	"github.com/hivepaas/shellwing-server/usecase/settings/imserviceuc/imservicedto"
	"github.com/hivepaas/shellwing-server/usecase/settings/keyauthuc/keyauthdto"
	"github.com/hivepaas/shellwing-server/usecase/settings/notificationuc/notificationdto"
	"github.com/hivepaas/shellwing-server/usecase/settings/oauthuc/oauthdto"
)

// Every list ListSetting serves reads the page asked for: a request that did
// not would be answered whole, whatever pageLimit says.
func TestEverySettingsListPages(t *testing.T) {
	for _, req := range []any{
		cloudstoragedto.NewListCloudStorageReq(),
		emaildto.NewListEmailReq(),
		imservicedto.NewListIMServiceReq(),
		keyauthdto.NewListKeyAuthReq(),
		notificationdto.NewListNotificationReq(),
		oauthdto.NewListOAuthReq(),
	} {
		_, ok := req.(interface{ PagingReq() *basedto.Paging })
		assert.True(t, ok, "%T does not page", req)
	}
}
