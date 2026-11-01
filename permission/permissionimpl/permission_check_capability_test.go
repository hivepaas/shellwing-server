package permissionimpl

import (
	"context"
	"testing"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/repository"
)

// The action is the check's own business. A caller that states none - which is
// what every call site now does - must still be answered, and one that states the
// wrong action must not be answered on those terms.
func TestCapabilityCheckSetsItsOwnAction(t *testing.T) {
	tests := []struct {
		name  string
		check *permission.CapabilityCheck
	}{
		{
			name:  "no action stated",
			check: &permission.CapabilityCheck{Capability: base.ResourceCapSecretReveal},
		},
		{
			name: "a wrong action stated is overwritten",
			check: &permission.CapabilityCheck{
				BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeDelete},
				Capability:      base.ResourceCapSecretReveal,
			},
		},
		{
			name: "AllOf and AnyOf are cleared too, so IsValid stays true",
			check: &permission.CapabilityCheck{
				BaseAccessCheck: permission.BaseAccessCheck{
					AllOf: []base.ActionType{base.ActionTypeRead},
					AnyOf: []base.ActionType{base.ActionTypeWrite},
				},
				Capability: base.ResourceCapSecretReveal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeCapabilityCheck(tt.check)

			if tt.check.Action != base.ActionTypeExecute {
				t.Errorf("Action = %q, want %q", tt.check.Action, base.ActionTypeExecute)
			}
			if len(tt.check.AllOf) != 0 || len(tt.check.AnyOf) != 0 {
				t.Error("AllOf and AnyOf must be cleared")
			}
			// Exactly one of the three has to be set, or VerifyAuth rejects it.
			if !tt.check.IsValid() {
				t.Error("the normalized check must be valid")
			}
		})
	}
}

// The action the check asks for and the action a grant is written with have to be
// the same one, or every grant is written and then never matches.
func TestCapabilityActionMatchesTheGrantedAction(t *testing.T) {
	granted := base.AccessActions{Exec: true}
	if !granted.Allows(capabilityAction) {
		t.Errorf("a capability grant carries %+v, which does not allow %q",
			granted, capabilityAction)
	}
}

// fakeACLPermissionRepo answers ListByResources from a fixed set of rows, filtered
// the way the real query filters them.
type fakeACLPermissionRepo struct {
	repository.ACLPermissionRepo
	rows []*entity.ACLPermission
}

func (f *fakeACLPermissionRepo) ListByResources(_ context.Context, _ database.IDB,
	resources []*base.PermissionResource, _ ...bunex.SelectQueryOption) ([]*entity.ACLPermission, error) {
	var res []*entity.ACLPermission
	for _, row := range f.rows {
		for _, r := range resources {
			if row.SubjectType == r.SubjectType && row.SubjectID == r.SubjectID &&
				row.ResourceType == r.ResourceType && row.ResourceID == r.ResourceID {
				res = append(res, row)
			}
		}
	}
	return res, nil
}

func memberAuth(userID string) *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: userID, Role: base.UserRoleMember},
	}}
}

// A capability is only grantable to a non-admin if CheckAccess actually reads the
// grant. A check type it does not know is answered false, so a missing case here
// made every capability an admin-only one, whatever was stored.
func TestCheckAccessAnswersCapabilityFromTheGrant(t *testing.T) {
	grant := &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		SubjectID:    "usr_holder",
		ResourceType: base.ResourceTypeCapability,
		ResourceID:   string(base.ResourceCapSecretReveal),
		Actions:      base.AccessActions{Exec: true},
	}
	p := &manager{aclPermissionRepo: &fakeACLPermissionRepo{rows: []*entity.ACLPermission{grant}}}

	tests := []struct {
		name string
		auth *basedto.Auth
		want bool
	}{
		{name: "a member holding the capability", auth: memberAuth("usr_holder"), want: true},
		{name: "a member without it", auth: memberAuth("usr_other"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.CheckAccess(context.Background(), nil, tt.auth, &permission.CapabilityCheck{
				Capability: base.ResourceCapSecretReveal,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("CheckAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A row for the capability that does not carry the execute action grants nothing.
func TestCheckAccessCapabilityNeedsTheExecuteAction(t *testing.T) {
	row := &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		SubjectID:    "usr_1",
		ResourceType: base.ResourceTypeCapability,
		ResourceID:   string(base.ResourceCapSecretReveal),
		Actions:      base.AccessActions{Read: true, Write: true, Del: true},
	}
	p := &manager{aclPermissionRepo: &fakeACLPermissionRepo{rows: []*entity.ACLPermission{row}}}

	got, err := p.CheckAccess(context.Background(), nil, memberAuth("usr_1"), &permission.CapabilityCheck{
		Capability: base.ResourceCapSecretReveal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("a row without the execute action must not grant the capability")
	}
}
