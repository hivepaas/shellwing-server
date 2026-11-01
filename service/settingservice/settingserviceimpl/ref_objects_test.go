package settingserviceimpl

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/repository"
)

var quotedID = regexp.MustCompile(`'([^']+)'`)

// fakeSettingRepo holds settings by ID and answers the IDs a query asks for, read
// from the SQL the query renders.
type fakeSettingRepo struct {
	repository.SettingRepo
	settings map[string]*entity.Setting
	queries  int
}

func (f *fakeSettingRepo) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	f.queries++
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	var found []*entity.Setting
	for _, m := range quotedID.FindAllStringSubmatch(sql, -1) {
		if setting := f.settings[m[1]]; setting != nil {
			found = append(found, setting)
		}
	}
	return found, nil, nil
}

func setting(t *testing.T, id string, typ base.SettingType, data entity.SettingData) *entity.Setting {
	t.Helper()
	s := &entity.Setting{ID: id, Type: typ, Status: base.SettingStatusActive}
	s.MustSetData(data)
	return s
}

// loadWithin fails the test when loading does not end: it used to recurse
// forever.
func loadWithin(t *testing.T, svc *service, refIDs *entity.RefObjectIDs, skipMissing bool) *entity.RefObjects {
	t.Helper()
	var refObjects *entity.RefObjects
	done := make(chan error, 1)
	go func() {
		if skipMissing {
			done <- svc.LoadRefObjectsByIDsSkipMissing(context.Background(), nil, &refObjects, nil, false, refIDs)
		} else {
			done <- svc.LoadRefObjectsByIDs(context.Background(), nil, &refObjects, nil, false, refIDs)
		}
	}()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("loading the references did not end")
	}
	return refObjects
}

// A storage and its key auth are loaded whole, along with what else was asked
// for, and loading ends.
func TestLoadRefObjectsFollowsTheReferences(t *testing.T) {
	repo := &fakeSettingRepo{settings: map[string]*entity.Setting{
		"cs1": setting(t, "cs1", base.SettingTypeCloudStorage,
			&entity.CloudStorage{S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "ka1"}}}),
		"ka1": setting(t, "ka1", base.SettingTypeKeyAuth, &entity.KeyAuth{}),
		"cs2": setting(t, "cs2", base.SettingTypeCloudStorage,
			&entity.CloudStorage{S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "ka1"}}}),
	}}

	refObjects := loadWithin(t, &service{settingRepo: repo},
		&entity.RefObjectIDs{RefSettingIDs: []string{"cs1", "cs2"}}, false)

	assert.Contains(t, refObjects.RefSettings, "cs1")
	assert.Contains(t, refObjects.RefSettings, "cs2")
	assert.Contains(t, refObjects.RefSettings, "ka1")
	assert.LessOrEqual(t, repo.queries, 2, "each setting is asked for once")
}

// References to settings that are gone are asked for once, and loading ends.
func TestLoadRefObjectsSkippingMissingOnesEnds(t *testing.T) {
	repo := &fakeSettingRepo{settings: map[string]*entity.Setting{
		"cs1": setting(t, "cs1", base.SettingTypeCloudStorage,
			&entity.CloudStorage{S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "gone-ka"}}}),
	}}

	refObjects := loadWithin(t, &service{settingRepo: repo},
		&entity.RefObjectIDs{RefSettingIDs: []string{"cs1", "gone-cs"}}, true)

	assert.Contains(t, refObjects.RefSettings, "cs1")
	assert.NotContains(t, refObjects.RefSettings, "gone-ka")
}
