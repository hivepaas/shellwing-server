package useruc

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

// spyPermissionManager records whether it was asked, and lets a test stand in for
// an admin (every current row is replaceable) or refuse the change outright.
type spyPermissionManager struct {
	permission.Manager
	called      bool
	desired     []*entity.ACLPermission
	replaceable []*entity.ACLPermission
	err         error
}

func (s *spyPermissionManager) AuthorizeAccessChanges(_ context.Context, _ database.IDB,
	_ *basedto.Auth, desired, current []*entity.ACLPermission) ([]*entity.ACLPermission, error) {
	s.called = true
	s.desired = desired
	if s.err != nil {
		return nil, s.err
	}
	if s.replaceable != nil {
		return s.replaceable, nil
	}
	return current, nil // what an admin gets back
}

func capabilityRow(userID string) *entity.ACLPermission {
	return &entity.ACLPermission{
		SubjectType:  base.SubjectTypeUser,
		SubjectID:    userID,
		ResourceType: base.ResourceTypeCapability,
		ResourceID:   string(base.ResourceCapSecretReveal),
		Actions:      base.AccessActions{Exec: true},
	}
}

// A request carrying only `capabilities` used to produce an empty resource-type
// list, and authorizeAccessChanges returns early on an empty list - so the grant
// was written without anybody being asked whether the caller may hand it out.
func TestCapabilityOnlyUpdateIsAuthorized(t *testing.T) {
	spy := &spyPermissionManager{err: errors.New("refused")}
	uc := &UC{permissionManager: spy}

	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserCapabilities(&entity.User{ID: "usr_target"},
		basedto.CapabilitySliceReq{base.ResourceCapSecretReveal},
		timeutil.NowUTC(), persistingData)

	// Exactly what UpdateUser passes when the request carries only capabilities.
	resourceTypes := accessResourceTypesToReplace(false, true)

	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		&entity.User{ID: "usr_target"}, resourceTypes, persistingData)

	if !spy.called {
		t.Fatal("the grant was applied without any authorization check")
	}
	if err == nil {
		t.Fatal("a refused change must fail the update")
	}
	if len(spy.desired) != 1 || spy.desired[0].ResourceType != base.ResourceTypeCapability {
		t.Errorf("the capability row must be submitted for checking, got %+v", spy.desired)
	}
}

// Dropping a capability from the list has to actually take it away.
func TestCapabilityIsRevocable(t *testing.T) {
	spy := &spyPermissionManager{}
	uc := &UC{permissionManager: spy}

	target := &entity.User{
		ID:       "usr_target",
		Accesses: []*entity.ACLPermission{capabilityRow("usr_target")},
	}

	// An empty-but-present list: "the user should hold no capabilities".
	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserCapabilities(target, basedto.CapabilitySliceReq{},
		timeutil.NowUTC(), persistingData)
	if len(persistingData.UpsertingAccesses) != 0 {
		t.Fatal("an empty list must write no grant")
	}

	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		target, accessResourceTypesToReplace(false, true), persistingData)
	if err != nil {
		t.Fatal(err)
	}

	if len(persistingData.DeletingAccesses) != 1 {
		t.Fatalf("the existing capability must be queued for removal, got %d rows",
			len(persistingData.DeletingAccesses))
	}
	if got := persistingData.DeletingAccesses[0]; got.ResourceType != base.ResourceTypeCapability ||
		got.ResourceID != string(base.ResourceCapSecretReveal) {
		t.Errorf("wrong row queued for removal: %+v", got)
	}
}

// Rows the actor cannot reach must survive an update rather than be swept away by
// a wholesale replace - the same rule the other grant types follow.
func TestCapabilityOutOfReachIsLeftAlone(t *testing.T) {
	spy := &spyPermissionManager{replaceable: []*entity.ACLPermission{}}
	uc := &UC{permissionManager: spy}

	target := &entity.User{
		ID:       "usr_target",
		Accesses: []*entity.ACLPermission{capabilityRow("usr_target")},
	}

	persistingData := &userservice.PersistingUserData{}
	err := uc.authorizeAccessChanges(context.Background(), nil, &basedto.Auth{},
		target, accessResourceTypesToReplace(false, true), persistingData)
	if err != nil {
		t.Fatal(err)
	}
	if len(persistingData.DeletingAccesses) != 0 {
		t.Error("a row the actor cannot revoke must not be removed")
	}
}

func TestPreparePersistingUserCapabilities(t *testing.T) {
	uc := &UC{}
	persistingData := &userservice.PersistingUserData{}

	uc.preparePersistingUserCapabilities(&entity.User{ID: "usr_1"},
		basedto.CapabilitySliceReq{base.ResourceCapSecretReveal},
		timeutil.NowUTC(), persistingData)

	if len(persistingData.UpsertingAccesses) != 1 {
		t.Fatalf("expected 1 ACL row, got %d", len(persistingData.UpsertingAccesses))
	}
	row := persistingData.UpsertingAccesses[0]
	if row.ResourceType != base.ResourceTypeCapability {
		t.Errorf("ResourceType = %q", row.ResourceType)
	}
	if row.SubjectType != base.SubjectTypeUser || row.SubjectID != "usr_1" {
		t.Errorf("wrong subject: %s/%s", row.SubjectType, row.SubjectID)
	}
	// The stored action has to be the one checkCapability asks for, or the grant
	// is written and then never matches.
	if !row.Actions.Allows(base.ActionTypeExecute) {
		t.Errorf("the row must carry the execute action, got %+v", row.Actions)
	}
}

func capabilityRows(rows []*entity.ACLPermission) []*entity.ACLPermission {
	var res []*entity.ACLPermission
	for _, row := range rows {
		if row.ResourceType == base.ResourceTypeCapability {
			res = append(res, row)
		}
	}
	return res
}

// An invite stores the capabilities it was sent with, next to the module accesses.
func TestInviteStoresCapabilities(t *testing.T) {
	uc := &UC{}
	persistingData := &userservice.PersistingUserData{}

	uc.preparePersistingUserInviteData(&userdto.InviteUserReq{
		Role:           base.UserRoleMember,
		SecurityOption: base.UserSecurityPassword2FA,
		ModuleAccesses: basedto.ModuleAccessSliceReq{{
			ModuleIDReq: basedto.ModuleIDReq{ID: string(base.ResourceModuleSettings)},
			Access:      base.AccessActions{Read: true},
		}},
		Capabilities: basedto.CapabilitySliceReq{base.ResourceCapSecretReveal},
	}, &userInviteData{User: &entity.User{ID: "usr_new"}}, persistingData)

	caps := capabilityRows(persistingData.UpsertingAccesses)
	if len(caps) != 1 || caps[0].ResourceID != string(base.ResourceCapSecretReveal) ||
		caps[0].SubjectID != "usr_new" {
		t.Fatalf("the invite must store the capability, got %+v", caps)
	}
	if len(persistingData.UpsertingAccesses) != 2 {
		t.Errorf("want the module access and the capability, got %d rows",
			len(persistingData.UpsertingAccesses))
	}
}

// An update replaces the capabilities when the request carries them, an empty
// list included, and leaves them alone when it does not.
func TestUpdateCapabilitiesOnlyWhenSent(t *testing.T) {
	tests := []struct {
		name     string
		caps     basedto.CapabilitySliceReq
		wantRows int
	}{
		{name: "absent leaves them alone", caps: nil, wantRows: 0},
		{name: "empty takes them all away", caps: basedto.CapabilitySliceReq{}, wantRows: 0},
		{name: "present writes them", caps: basedto.CapabilitySliceReq{base.ResourceCapSecretReveal}, wantRows: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &UC{}
			persistingData := &userservice.PersistingUserData{}
			req := &userdto.UpdateUserReq{ID: "usr_1", Capabilities: tt.caps}

			uc.prepareUpdatingUserData(req, &userUpdateData{User: &entity.User{ID: "usr_1"}}, persistingData)

			if got := len(capabilityRows(persistingData.UpsertingAccesses)); got != tt.wantRows {
				t.Errorf("want %d capability rows, got %d", tt.wantRows, got)
			}
			// Whether the existing rows are replaced is decided by the resource
			// types handed to authorizeAccessChanges, which is what UpdateUser
			// derives from the same field.
			replaced := slices.Contains(accessResourceTypesToReplace(false, req.Capabilities != nil),
				base.ResourceTypeCapability)
			if replaced != (tt.caps != nil) {
				t.Errorf("capabilities replaced = %v, want %v", replaced, tt.caps != nil)
			}
		})
	}
}
