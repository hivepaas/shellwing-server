package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
)

/// Fakes

// fakePermissionManager stands in for the gate, which is tested where it lives -
// permissionimpl. What matters here is whether this layer asks at all, and what
// it says the secret is when it does.
type fakePermissionManager struct {
	permission.Manager
	err      error
	subjects []*permission.RevealSubject
}

func (f *fakePermissionManager) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	f.subjects = append(f.subjects, subject)
	return f.err
}

/// Helpers

func newRevealUC(refusal error) (*BaseUC, *fakePermissionManager) {
	perms := &fakePermissionManager{err: refusal}
	return &BaseUC{PermissionManager: perms}, perms
}

// newEmailSetting builds a stored setting whose password is encrypted, the state
// a setting is in when it comes back from the database.
func newEmailSetting(t *testing.T) *entity.Setting {
	t.Helper()
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)

	setting := &entity.Setting{
		ID:   "set_1",
		Name: "smtp-main",
		Type: base.SettingTypeEmail,
	}
	if err := setting.SetData(&entity.Email{SMTP: &entity.EmailSMTP{
		Host:     "smtp.example.com",
		Username: "admin",
		Password: entity.NewEncryptedField("the-real-password"),
	}}); err != nil {
		t.Fatal(err)
	}
	// Round-trip through the stored form so the field holds ciphertext only.
	stored := &entity.Setting{ID: setting.ID, Name: setting.Name, Type: setting.Type, Data: setting.Data}
	stored.ObjectID, stored.CurrentObjectID = "obj_1", "obj_1"
	return stored
}

func storedPassword(t *testing.T, setting *entity.Setting) *entity.EncryptedField {
	t.Helper()
	email, err := setting.AsEmail()
	if err != nil {
		t.Fatal(err)
	}
	return &email.SMTP.Password
}

// revealScope is the scope the reveal is being asked for. It is what the audit
// entry is filed under, and an entry filed under the wrong scope is invisible to
// the scoped audit log listing - which is where somebody would go looking for it.
var revealScope = &entity.ObjectScope{ScopeType: base.ObjectScopeUser, UserID: "obj_1"}

/// Tests

func TestRevealSecretsDecryptsWhenAuthorized(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newEmailSetting(t)

	_, err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if err != nil {
		t.Fatalf("reveal must succeed: %v", err)
	}

	if got := storedPassword(t, setting).String(); got != "the-real-password" {
		t.Errorf("secret was not revealed, got %q", got)
	}
	if len(perms.subjects) != 1 {
		t.Fatalf("want one authorization, got %d", len(perms.subjects))
	}
}

// The subject is what the audit entry is built from, so getting it wrong here
// files the record under something nobody will search for.
func TestRevealSecretsNamesTheSettingItIsAbout(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newEmailSetting(t)

	_, _ = uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)

	subject := perms.subjects[0]
	if subject.Scope != base.ObjectScopeUser || subject.ObjectID != "obj_1" {
		t.Errorf("asked under %v/%q, want the scope the reveal was for", subject.Scope, subject.ObjectID)
	}
	if subject.ResType != base.ResourceTypeSetting || subject.ResID != "set_1" || subject.ResName != "smtp-main" {
		t.Errorf("subject = %q/%q/%q", subject.ResType, subject.ResID, subject.ResName)
	}
	if subject.Source != base.AuditLogSourceAPIGet {
		t.Errorf("source = %q", subject.Source)
	}
}

func TestRevealSecretsLeavesTheSecretEncryptedWhenRefused(t *testing.T) {
	refusal := errors.New("not allowed")
	uc, _ := newRevealUC(refusal)
	setting := newEmailSetting(t)

	_, err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)

	if !errors.Is(err, refusal) {
		t.Fatalf("the refusal must come back, got %v", err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("the secret must stay encrypted when the caller is refused")
	}
}

func TestRevealSecretsNotRequested(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newEmailSetting(t)

	_, err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, false, setting)
	if err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("nothing may be decrypted when reveal was not asked for")
	}
	if len(perms.subjects) != 0 {
		t.Error("an ordinary read must not cost an authorization or an audit row")
	}
}

// An inherited setting is read through this scope, not owned by it.
func TestRevealSecretsInheritedIsNeverRevealed(t *testing.T) {
	uc, perms := newRevealUC(nil)
	setting := newEmailSetting(t)
	setting.CurrentObjectID = "obj_2"

	_, err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if err != nil {
		t.Fatal(err)
	}
	if !storedPassword(t, setting).IsEncrypted() {
		t.Error("an inherited setting must not hand out its secrets")
	}
	if len(perms.subjects) != 0 {
		t.Error("nothing was revealed, so there is nothing to authorize")
	}
}

// A type holding no secrets must not cost a permission check or a row: a cloud
// storage's key is its linked key auth's.
func TestRevealSecretsOnTypeWithoutSecrets(t *testing.T) {
	uc, perms := newRevealUC(errors.New("would refuse if asked"))

	setting := &entity.Setting{ID: "set_2", Name: "bucket", Type: base.SettingTypeCloudStorage}
	if err := setting.SetData(&entity.CloudStorage{S3: &entity.CloudStorageS3{Bucket: "b"}}); err != nil {
		t.Fatal(err)
	}
	setting.ObjectID, setting.CurrentObjectID = "obj_1", "obj_1"

	_, err := uc.revealSecrets(context.Background(), nil, &basedto.Auth{}, revealScope, true, setting)
	if err != nil {
		t.Fatalf("a type without secrets must not be refused: %v", err)
	}
	if len(perms.subjects) != 0 {
		t.Error("nothing was revealed, so there is nothing to authorize")
	}
}
