package userdto

import (
	"testing"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
)

// Both requests that write capabilities validate them against the server's own
// list - the API-key one is not among them, the server having no API keys.
func TestUserReqsValidateCapabilities(t *testing.T) {
	tests := []struct {
		name    string
		caps    basedto.CapabilitySliceReq
		wantErr bool
	}{
		{name: "absent", caps: nil},
		{name: "empty", caps: basedto.CapabilitySliceReq{}},
		{name: "reveal secrets", caps: basedto.CapabilitySliceReq{base.ResourceCapSecretReveal}},
		{name: "unknown", caps: basedto.CapabilitySliceReq{"cap::api-key::create"}, wantErr: true},
		{
			name:    "duplicated",
			caps:    basedto.CapabilitySliceReq{base.ResourceCapSecretReveal, base.ResourceCapSecretReveal},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invite := &InviteUserReq{
				Email:          "someone@example.com",
				Role:           base.UserRoleMember,
				SecurityOption: base.UserSecurityPassword2FA,
				Capabilities:   tt.caps,
			}
			update := &UpdateUserReq{ID: "01JAB9XED0GTXBSQDFVYAJ8WA1", Capabilities: tt.caps}

			for reqName, errs := range map[string]int{
				"invite": len(invite.Validate()),
				"update": len(update.Validate()),
			} {
				if tt.wantErr && errs == 0 {
					t.Errorf("%s: expected a validation error, got none", reqName)
				}
				if !tt.wantErr && errs > 0 {
					t.Errorf("%s: expected no validation error, got %d", reqName, errs)
				}
			}
		})
	}
}
