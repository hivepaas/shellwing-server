package licensedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/service/licenseservice"
)

// maxKeyLen bounds the text taken as a key: a key is a few kilobytes.
const maxKeyLen = 32 << 10

// LicenseResp is the license in force, and the seats taken. MaxUsers -1: no limit. FreeUsers is the limit with no
// license in force, which is back when one ends.
type LicenseResp struct {
	Status         string    `json:"status"`
	InstallationID string    `json:"installationId"`
	Plan           string    `json:"plan"`
	MaxUsers       int       `json:"maxUsers"`
	FreeUsers      int       `json:"freeUsers"`
	UsedUsers      int       `json:"usedUsers"`
	Trial          bool      `json:"trial"`
	CustomerName   string    `json:"customerName,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt,omitzero"`
	GraceEndsAt    time.Time `json:"graceEndsAt,omitzero"`
}

func TransformLicense(state *licenseservice.State, usedUsers int) *LicenseResp {
	resp := &LicenseResp{
		Status:         string(state.Status),
		InstallationID: state.InstallationID,
		Plan:           state.Plan,
		MaxUsers:       state.MaxUsers,
		FreeUsers:      state.FreeUsers,
		UsedUsers:      usedUsers,
		GraceEndsAt:    state.GraceEndsAt,
	}
	if state.Key != nil {
		resp.Trial = state.Key.Trial
		resp.CustomerName = state.Key.Customer.Name
		resp.ExpiresAt = state.Key.ExpiresAt
	}
	return resp
}

type GetLicenseResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *LicenseResp  `json:"data"`
}

// InstallLicenseReq carries a key as it was pasted: wrapped or indented is the same key.
type InstallLicenseReq struct {
	Key string `json:"key"`
}

func NewInstallLicenseReq() *InstallLicenseReq {
	return &InstallLicenseReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *InstallLicenseReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.Key, true, 1, maxKeyLen, "key")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type InstallLicenseResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *LicenseResp  `json:"data"`
}

type RemoveLicenseResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *LicenseResp  `json:"data"`
}
