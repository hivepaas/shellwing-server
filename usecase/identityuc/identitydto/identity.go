package identitydto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
)

// IdentityResp is a person's identity: the public keys, base64, and the fingerprint people compare.
type IdentityResp struct {
	X25519Public  []byte    `json:"x25519Public"`
	Ed25519Public []byte    `json:"ed25519Public"`
	Signature     []byte    `json:"signature"`
	Fingerprint   string    `json:"fingerprint"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func TransformIdentity(i *entity.UserIdentity) *IdentityResp {
	if i == nil {
		return nil
	}
	return &IdentityResp{
		X25519Public:  i.X25519Public,
		Ed25519Public: i.Ed25519Public,
		Signature:     i.Signature,
		Fingerprint:   i.Fingerprint,
		UpdatedAt:     i.UpdatedAt,
	}
}

// GetIdentityResp has no data when the person's app has published no identity yet.
type GetIdentityResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *IdentityResp `json:"data"`
}

// PutIdentityReq is the identity the app publishes: the public keys and the signature, base64.
type PutIdentityReq struct {
	X25519Public  []byte `json:"x25519Public"`
	Ed25519Public []byte `json:"ed25519Public"`
	Signature     []byte `json:"signature"`
}

func NewPutIdentityReq() *PutIdentityReq {
	return &PutIdentityReq{}
}

// Validate implements interface basedto.ReqValidator. Whether the keys are an identity is the use case's to check.
func (req *PutIdentityReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateRequiredField(req.X25519Public, "x25519Public")...,
	))
}

type PutIdentityResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *IdentityResp `json:"data"`
}
