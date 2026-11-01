package settings

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

/// Helpers

// fakeAuditService records what it was handed. It lives here because the audit
// entry is what these tests are about; the reveal path writes its entries from
// permissionimpl now, and its tests keep their own copy.
type fakeAuditService struct {
	auditservice.Service
	entries []*auditservice.Entry
	err     error
}

func (f *fakeAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	f.entries = append(f.entries, entry)
	return f.err
}

func newAuditUC(auditErr error) (*BaseUC, *fakeAuditService) {
	audit := &fakeAuditService{err: auditErr}
	return &BaseUC{AuditService: audit}, audit
}

var auditScope = &entity.ObjectScope{ScopeType: base.ObjectScopeUser, UserID: "obj_1"}

func newAuditReq() *BaseSettingReq {
	return &BaseSettingReq{
		Type:  base.SettingTypeEmail,
		Scope: auditScope,
		Auth:  &basedto.Auth{},
	}
}

// auditDetail mirrors the JSON the builder renders, so the tests assert on the
// stored shape rather than on an internal struct.
type auditDetail struct {
	SettingType string                        `json:"settingType"`
	Status      string                        `json:"status"`
	Changes     map[string]auditdetail.Change `json:"changes"`
	DataFields  []string                      `json:"dataFields"`
}

func decodeDetail(t *testing.T, detail string) *auditDetail {
	t.Helper()
	out := &auditDetail{}
	if err := json.Unmarshal([]byte(detail), out); err != nil {
		t.Fatalf("detail is not valid JSON: %v", err)
	}
	return out
}

/// Tests

func TestRecordSettingAuditFilesTheEntry(t *testing.T) {
	uc, audit := newAuditUC(nil)
	old := &entity.Setting{ID: "set_1", Name: "smtp-main", Type: base.SettingTypeEmail}
	current := &entity.Setting{ID: "set_1", Name: "smtp-main-2", Type: base.SettingTypeEmail}

	err := uc.recordSettingAudit(context.Background(), nil, newAuditReq(),
		base.AuditLogTypeSettingUpdate, base.AuditLogSourceAPIUpdate, old, current)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(audit.entries))
	}

	entry := audit.entries[0]
	if entry.Type != base.AuditLogTypeSettingUpdate {
		t.Errorf("type = %q", entry.Type)
	}
	if entry.Result != base.AuditLogResultAllowed {
		t.Errorf("result = %q", entry.Result)
	}
	// The scope decides which scoped listing the entry shows up in, so an entry
	// filed under the wrong one is invisible where somebody would look for it.
	if entry.Scope != base.ObjectScopeUser || entry.ObjectID != "obj_1" {
		t.Errorf("scope = %q/%q", entry.Scope, entry.ObjectID)
	}
	if entry.ResType != base.ResourceType(base.SettingTypeEmail) {
		t.Errorf("resType = %q", entry.ResType)
	}
	// The name is snapshotted as it reads after the change, not before.
	if entry.ResID != "set_1" || entry.ResName != "smtp-main-2" {
		t.Errorf("res = %q/%q", entry.ResID, entry.ResName)
	}
}

// The whole flow is expected to abort when the record cannot be written, so the
// error has to come back rather than be swallowed.
func TestRecordSettingAuditPropagatesFailure(t *testing.T) {
	uc, _ := newAuditUC(errors.New("database is down"))
	setting := &entity.Setting{ID: "set_1", Name: "smtp-main", Type: base.SettingTypeEmail}

	err := uc.recordSettingAudit(context.Background(), nil, newAuditReq(),
		base.AuditLogTypeSettingDelete, base.AuditLogSourceAPIDelete, setting, nil)
	if err == nil {
		t.Fatal("want an error when the record cannot be written")
	}
}

// A usecase that forgets to set req.Auth must break loudly. The alternative is an
// entry naming nobody, which answers none of the questions it exists for.
func TestRecordSettingAuditRefusesWithoutAuth(t *testing.T) {
	uc, audit := newAuditUC(nil)
	req := newAuditReq()
	req.Auth = nil
	setting := &entity.Setting{ID: "set_1", Name: "smtp-main", Type: base.SettingTypeEmail}

	err := uc.recordSettingAudit(context.Background(), nil, req,
		base.AuditLogTypeSettingUpdate, base.AuditLogSourceAPIUpdate, setting, setting)
	if err == nil {
		t.Fatal("want an error when the caller is unknown")
	}
	if len(audit.entries) != 0 {
		t.Errorf("want no entry written, got %d", len(audit.entries))
	}
}

func TestSettingAuditDetailListsEnvelopeChanges(t *testing.T) {
	expiry := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	old := &entity.Setting{
		Type: base.SettingTypeEmail, Name: "smtp-main",
		Status: base.SettingStatusActive, Inheritable: false,
	}
	current := &entity.Setting{
		Type: base.SettingTypeEmail, Name: "smtp-main",
		Status: base.SettingStatusDisabled, Inheritable: true, ExpireAt: expiry,
	}

	detail := decodeDetail(t, settingAuditDetail(current, old, current))
	if detail.SettingType != string(base.SettingTypeEmail) {
		t.Errorf("settingType = %q", detail.SettingType)
	}

	if _, ok := detail.Changes["name"]; ok {
		t.Error("name did not change and must not be listed")
	}
	for _, field := range []string{"status", "inheritable", "expireAt"} {
		if _, ok := detail.Changes[field]; !ok {
			t.Errorf("%q changed but is not listed", field)
		}
	}
	if got := detail.Changes["expireAt"].To; got != expiry.Format(time.RFC3339) {
		t.Errorf("expireAt.to = %v", got)
	}
}

// The one thing the detail must never contain. Setting.Data holds each type's own
// payload, and for every type implementing secretDecrypter part of it is a
// secret; an audit row outlives the setting it describes.
func TestSettingAuditDetailNeverCarriesSettingData(t *testing.T) {
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)

	const password = "the-real-password"
	current := &entity.Setting{ID: "set_1", Name: "smtp-main", Type: base.SettingTypeEmail}
	if err := current.SetData(&entity.Email{SMTP: &entity.EmailSMTP{
		Username: "admin",
		Password: entity.NewEncryptedField(password),
	}}); err != nil {
		t.Fatal(err)
	}
	old := &entity.Setting{ID: "set_1", Name: "smtp-main-old", Type: base.SettingTypeEmail}

	detail := settingAuditDetail(current, old, current)
	if strings.Contains(detail, password) {
		t.Fatal("detail carries the plaintext secret")
	}
	// Ciphertext is no better: it is still the secret, and the row outlives the
	// key rotation that would otherwise retire it.
	if current.Data != "" && strings.Contains(detail, current.Data) {
		t.Fatal("detail carries the stored payload")
	}
	if strings.Contains(detail, "admin") {
		t.Fatal("detail carries a payload field")
	}
}

func TestSettingAuditDetailOmitsChangesOnCreateAndDelete(t *testing.T) {
	setting := &entity.Setting{
		Type: base.SettingTypeEmail, Name: "smtp-main", Status: base.SettingStatusActive,
	}

	created := decodeDetail(t, settingAuditDetail(setting, nil, setting))
	if created.Changes != nil {
		t.Errorf("a create has nothing to diff, got %v", created.Changes)
	}
	if created.Status != string(base.SettingStatusActive) {
		t.Errorf("status = %q", created.Status)
	}

	deleted := decodeDetail(t, settingAuditDetail(setting, setting, nil))
	if deleted.Changes != nil {
		t.Errorf("a delete has nothing to diff, got %v", deleted.Changes)
	}
}

/// Payload field diff

func newEmailSMTP(t *testing.T, password string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "set_1", Name: "smtp-main", Type: base.SettingTypeEmail}
	if err := setting.SetData(&entity.Email{SMTP: &entity.EmailSMTP{
		Username: "admin",
		Password: entity.NewEncryptedField(password),
	}}); err != nil {
		t.Fatal(err)
	}
	// Round-trip through the stored form, which is how a setting arrives from the
	// database and therefore how the flows hand it to the audit.
	return &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
}

func withDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)
}

func TestSettingDataFieldsChangedNamesTheField(t *testing.T) {
	withDataKey(t)
	old := newEmailSMTP(t, "one")
	current := newEmailSMTP(t, "two")

	fields := decodeDetail(t, settingAuditDetail(current, old, current)).DataFields
	if len(fields) != 1 || fields[0] != "smtp.password" {
		t.Fatalf("want [smtp.password], got %v", fields)
	}
}

// The reason the walk is typed. Sealing the same plaintext again picks a fresh
// nonce, so the stored JSON differs while nothing actually changed; a textual
// comparison would report a password change on every save that echoed it back.
func TestSettingDataFieldsChangedIgnoresResealedSecret(t *testing.T) {
	withDataKey(t)
	old := newEmailSMTP(t, "same")
	current := newEmailSMTP(t, "same")

	if old.Data == current.Data {
		t.Skip("ciphertext is deterministic here, the case cannot arise")
	}
	if fields := decodeDetail(t, settingAuditDetail(current, old, current)).DataFields; len(fields) != 0 {
		t.Fatalf("nothing changed, got %v", fields)
	}
}

func TestSettingDataFieldsChangedWalksNestedValues(t *testing.T) {
	withDataKey(t)
	build := func(fromName string) *entity.Setting {
		setting := &entity.Setting{ID: "set_1", Name: "mailer", Type: base.SettingTypeEmail}
		if err := setting.SetData(&entity.Email{HTTP: &entity.EmailHTTP{
			Endpoint: "https://mail.example.com/send",
			FieldMapping: &entity.EmailHTTPFieldMapping{
				FromAddress: "from",
				FromName:    fromName,
			},
		}}); err != nil {
			t.Fatal(err)
		}
		return &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
	}

	fields := decodeDetail(t, settingAuditDetail(build("sender"), build("from_name"), build("sender"))).DataFields
	if len(fields) != 1 || fields[0] != "http.fieldMapping.fromName" {
		t.Fatalf("want the nested path, got %v", fields)
	}
}

// The case the `hpenc:`/`hpsalt:` marker rule would miss: a plain string field
// holding whatever a user typed into it, which is routinely a credential.
func TestSettingAuditDetailNeverCarriesFreeTextPayload(t *testing.T) {
	withDataKey(t)
	const leaked = "AKIAIOSFODNN7EXAMPLE"
	build := func(apiKey string) *entity.Setting {
		setting := &entity.Setting{ID: "set_1", Name: "mailer", Type: base.SettingTypeEmail}
		if err := setting.SetData(&entity.Email{HTTP: &entity.EmailHTTP{
			Endpoint: "https://mail.example.com/send",
			// A header value is a plain string holding whatever was typed into it.
			Headers: map[string]string{"X-Api-Key": apiKey},
		}}); err != nil {
			t.Fatal(err)
		}
		return &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
	}
	old, current := build("old"), build(leaked)

	detail := settingAuditDetail(current, old, current)
	if strings.Contains(detail, leaked) {
		t.Fatal("detail carries free-text payload content")
	}
	// The field is still named, which is the whole point of naming and not valuing.
	if !strings.Contains(detail, "http.headers.X-Api-Key") {
		t.Fatalf("the changed field is not named: %s", detail)
	}
}
