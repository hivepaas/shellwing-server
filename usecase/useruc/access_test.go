package useruc

import (
	"slices"
	"testing"

	"github.com/hivepaas/shellwing-server/base"
)

func TestAccessResourceTypesToReplace(t *testing.T) {
	tests := []struct {
		name                string
		replaceModules      bool
		replaceCapabilities bool
		wantResourceType    []base.ResourceType
	}{
		{
			name: "nothing to replace when the request carries no access list",
		},
		{
			name:             "modules only",
			replaceModules:   true,
			wantResourceType: []base.ResourceType{base.ResourceTypeModule},
		},
		{
			// Leaving capabilities out made a capabilities-only update skip the
			// authorization check entirely, and made a granted capability
			// impossible to revoke.
			name:                "capabilities only",
			replaceCapabilities: true,
			wantResourceType:    []base.ResourceType{base.ResourceTypeCapability},
		},
		{
			name:                "both of them",
			replaceModules:      true,
			replaceCapabilities: true,
			wantResourceType: []base.ResourceType{
				base.ResourceTypeModule, base.ResourceTypeCapability,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessResourceTypesToReplace(tt.replaceModules, tt.replaceCapabilities)
			if !slices.Equal(got, tt.wantResourceType) {
				t.Errorf("accessResourceTypesToReplace() = %v, want %v", got, tt.wantResourceType)
			}
		})
	}
}
