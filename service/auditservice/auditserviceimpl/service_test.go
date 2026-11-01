package auditserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

type fakeAuditLogRepo struct {
	repository.AuditLogRepo
	inserted *entity.AuditLog
}

func (f *fakeAuditLogRepo) Insert(
	_ context.Context, _ database.IDB, auditLog *entity.AuditLog, _ ...bunex.InsertQueryOption,
) error {
	f.inserted = auditLog
	return nil
}

// Section is a column, not a field inside Detail, so it has to be carried across
// rather than left to whatever the caller happened to put in the JSON.
func TestRecordCarriesTheSectionIntoItsOwnColumn(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	svc := New(repo)

	err := svc.Record(context.Background(), nil, &auditservice.Entry{
		Type:    base.AuditLogTypeUserUpdate,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: "account",
		Result:  base.AuditLogResultAllowed,
		Detail:  `{"moduleAccessesReplaced":true}`,
	})

	assert.NoError(t, err)
	assert.Equal(t, "account", repo.inserted.Section)
	// The listing drops Detail once it passes a hundred characters, which is why
	// the section had to stop living in there.
	assert.NotContains(t, repo.inserted.Detail, "section")
}

// An entry for a type that covers one endpoint has no section, and an empty
// string must reach the column as NULL rather than as an indexed empty value.
func TestRecordLeavesTheSectionEmptyWhenThereIsNone(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	svc := New(repo)

	err := svc.Record(context.Background(), nil, &auditservice.Entry{
		Type:   base.AuditLogTypeSecretReveal,
		Scope:  base.ObjectScopeGlobal,
		Result: base.AuditLogResultAllowed,
	})

	assert.NoError(t, err)
	assert.Empty(t, repo.inserted.Section)
}
