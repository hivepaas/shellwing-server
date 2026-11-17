package auditserviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/auditsink"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

// fakeSink records what it was given, or refuses everything.
type fakeSink struct {
	fail    error
	msgIDs  []string
	payload []byte
}

func (f *fakeSink) Name() string { return "fake" }

func (f *fakeSink) Send(_ context.Context, msgID string, payload []byte) error {
	if f.fail != nil {
		return f.fail
	}
	f.msgIDs = append(f.msgIDs, msgID)
	f.payload = payload
	return nil
}

func (f *fakeSink) Close() error { return nil }

func useAudit(t *testing.T, audit config.Audit) {
	t.Helper()
	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })
	config.SetCurrent(&config.Config{Audit: audit})
}

func anEntry() *auditservice.Entry {
	return &auditservice.Entry{
		Type:    base.AuditLogTypeSSHCertIssue,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Result:  base.AuditLogResultAllowed,
		ResType: base.ResourceTypeSSHCert,
		ResID:   "123",
		Detail:  `{"orgId":"01JC"}`,
		Auth: &basedto.Auth{User: &basedto.User{
			User:       &entity.User{ID: "u1", Username: "alice"},
			AuthClaims: &jwtsession.AuthClaims{UID: "device-1", UserID: "u1"},
		}},
	}
}

// What goes on the wire is pinned here on purpose. A reader at the far end parses these names, and renaming a field in
// the entity would quietly change a format somebody else's alerts are written against.
func TestTheRecordReachesTheSinkAsWellAsTheDatabase(t *testing.T) {
	repo, sink := &fakeAuditLogRepo{}, &fakeSink{}
	useAudit(t, config.Audit{Required: true})
	svc := New(repo, sink, logging.GlobalLogger())

	require.NoError(t, svc.Record(context.Background(), nil, anEntry()))
	require.NotNil(t, repo.inserted, "the database first")
	require.Equal(t, []string{string(base.AuditLogTypeSSHCertIssue)}, sink.msgIDs,
		"and the kind of action is the message id, so a collector can route on it")

	var sent map[string]any
	require.NoError(t, json.Unmarshal(sink.payload, &sent))
	keys := make([]string, 0, len(sent))
	for key := range sent {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{
		"actorId", "actorName", "actorType", "createdAt", "detail", "id",
		"resId", "resType", "result", "scope", "sessionUid", "source", "type",
	}, keys, "the wire format: a rename here is a break for whoever reads these")

	assert.Equal(t, repo.inserted.ID, sent["id"], "the same record, not a second one")
	assert.Equal(t, "alice", sent["actorName"])
	assert.Equal(t, "device-1", sent["sessionUid"], "which session, so one stolen can be told from the owner's use")
	assert.Equal(t, `{"orgId":"01JC"}`, sent["detail"])
}

// The stance the package already takes about the database, now about the sink: an action that cannot be recorded does
// not happen. The caller aborts on the error, so the row rolls back with it.
func TestAnActionWhoseRecordCannotBeSentDoesNotHappen(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	sink := &fakeSink{fail: errors.New("connection refused")}
	useAudit(t, config.Audit{Required: true})
	svc := New(repo, sink, logging.GlobalLogger())

	err := svc.Record(context.Background(), nil, anEntry())
	require.ErrorIs(t, err, hperrors.ErrAuditSinkUnavailable)
	assert.Contains(t, err.Error(), "ERR_AUDIT_SINK_UNAVAILABLE")
}

// And the other way for an operator who would rather the server kept working. The record is still in the database;
// what is lost is the copy, which the log says.
func TestWithARecordNotRequiredTheActionStillHappens(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	sink := &fakeSink{fail: errors.New("connection refused")}
	useAudit(t, config.Audit{Required: false})
	svc := New(repo, sink, logging.GlobalLogger())

	require.NoError(t, svc.Record(context.Background(), nil, anEntry()))
	assert.NotNil(t, repo.inserted, "and the database still has it")
}

// An installation that configured no sink is not affected by any of this.
func TestWithNoSinkNothingChanges(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	useAudit(t, config.Audit{Sink: config.AuditSinkNone, Required: true})
	svc := New(repo, auditsink.Nop(), logging.GlobalLogger())

	require.NoError(t, svc.Record(context.Background(), nil, anEntry()))
	assert.NotNil(t, repo.inserted)
}
