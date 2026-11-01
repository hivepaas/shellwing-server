package userdto

import (
	"testing"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
)

func read() base.AccessActions { return base.AccessActions{Read: true} }

// The details response reports the module accesses and the capabilities, each
// from its own kind of ACL row, so the caller can render and edit both without a
// second lookup.
func TestTransformUserDetailsReportsModulesAndCapabilities(t *testing.T) {
	user := &entity.User{
		ID: "usr_1",
		Accesses: []*entity.ACLPermission{
			{ResourceType: base.ResourceTypeModule, ResourceID: string(base.ResourceModuleSettings), Actions: read()},
			{ResourceType: base.ResourceTypeCapability, ResourceID: string(base.ResourceCapSecretReveal),
				Actions: base.AccessActions{Exec: true}},
		},
	}

	resp, err := TransformUserDetails(user)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.ModuleAccesses) != 1 || resp.ModuleAccesses[0].ID != string(base.ResourceModuleSettings) ||
		!resp.ModuleAccesses[0].Access.Equal(read()) {
		t.Errorf("module accesses = %+v", resp.ModuleAccesses)
	}
	if len(resp.Capabilities) != 1 || resp.Capabilities[0] != base.ResourceCapSecretReveal {
		t.Errorf("capabilities = %v", resp.Capabilities)
	}
}

// A user loaded without their accesses - getAccesses not asked for - reports
// neither list rather than failing.
func TestTransformUserDetailsWithoutAccesses(t *testing.T) {
	resp, err := TransformUserDetails(&entity.User{ID: "usr_1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ModuleAccesses != nil || resp.Capabilities != nil {
		t.Errorf("want no accesses, got modules %+v and capabilities %v", resp.ModuleAccesses, resp.Capabilities)
	}
}
