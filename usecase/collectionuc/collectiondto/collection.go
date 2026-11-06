package collectiondto

import (
	"math"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/identityuc/identitydto"
)

const (
	nameMaxLen        = 100
	descriptionMaxLen = 500
	grantsMax         = 1000
	wrapsMax          = 1000
)

// CollectionResp is a collection as the person sees it: what they may do, and whether they hold its current key.
// Waiting and NeedsNewKey are for those who manage its access, and 0 and false for others.
type CollectionResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// KeyVersion is the version of the key its records are sealed with now: 0 until an app makes the first.
	KeyVersion int `json:"keyVersion"`
	// Permission is the person's: connect, edit or manage, or empty for an admin with no grant.
	Permission base.CollectionPermission `json:"permission"`
	HasKey     bool                      `json:"hasKey"`
	// Waiting counts the people with access who are not given the current key yet.
	Waiting int `json:"waiting"`
	// NeedsNewKey says somebody holding the current key has lost access: an app with Manage replaces it.
	NeedsNewKey bool `json:"needsNewKey"`
}

type ListCollectionsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data []*CollectionResp `json:"data"`
}

// GrantResp is a permission granted to a person or a group.
type GrantResp struct {
	UserID     string                    `json:"userId,omitempty"`
	Username   string                    `json:"username,omitempty"`
	FullName   string                    `json:"fullName,omitempty"`
	GroupID    string                    `json:"groupId,omitempty"`
	GroupName  string                    `json:"groupName,omitempty"`
	Permission base.CollectionPermission `json:"permission"`
}

func TransformGrant(g *entity.CollectionGrantee) *GrantResp {
	return &GrantResp{
		UserID:     g.UserID,
		Username:   g.Username,
		FullName:   g.FullName,
		GroupID:    g.GroupID,
		GroupName:  g.GroupName,
		Permission: g.Permission,
	}
}

// MemberResp is a person with access to the collection, through their role, their grant or their groups'.
type MemberResp struct {
	UserID     string                    `json:"userId"`
	Username   string                    `json:"username"`
	FullName   string                    `json:"fullName"`
	Role       base.OrgRole              `json:"role"`
	Permission base.CollectionPermission `json:"permission"`
	// Served is false for a suspended member or a disabled account: they keep their access, and get nothing.
	Served bool `json:"served"`
	HasKey bool `json:"hasKey"`
	// Identity is what the key is wrapped for, and the fingerprint to compare first. None until their app publishes
	// one: nobody can give them the key before.
	Identity *identitydto.IdentityResp `json:"identity"`
}

// CollectionDetailResp is a collection, with its grants and members for those who manage its access.
type CollectionDetailResp struct {
	Collection *CollectionResp `json:"collection"`
	Grants     []*GrantResp    `json:"grants"`
	Members    []*MemberResp   `json:"members"`
}

type GetCollectionResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *CollectionDetailResp `json:"data"`
}

// CollectionReq is a collection's name and description.
type CollectionReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewCollectionReq() *CollectionReq {
	return &CollectionReq{}
}

func (req *CollectionReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CollectionReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.Name, true, 1, nameMaxLen, "name")
	validators = append(validators, basedto.ValidateStr(&req.Description, false, 0, descriptionMaxLen,
		"description")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CollectionDataResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *CollectionResp `json:"data"`
}

type DeleteCollectionResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// GrantReq grants a person or a group: one of userId and groupId.
type GrantReq struct {
	UserID     string                    `json:"userId"`
	GroupID    string                    `json:"groupId"`
	Permission base.CollectionPermission `json:"permission"`
}

// SetGrantsReq makes these the collection's grants, and no others.
type SetGrantsReq struct {
	Grants []*GrantReq `json:"grants"`
}

func NewSetGrantsReq() *SetGrantsReq {
	return &SetGrantsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetGrantsReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateSliceEx(req.Grants, false, 0, grantsMax, nil, "grants")
	for _, g := range req.Grants {
		if g == nil {
			validators = append(validators, basedto.ValidateCond(false, "grants")...)
			continue
		}
		validators = append(validators, basedto.ValidateMutualExclusiveFields(
			(g.UserID == "") != (g.GroupID == ""), "userId", "groupId")...)
		validators = append(validators, basedto.ValidateStrIn(&g.Permission, true, base.AllCollectionPermissions,
			"permission")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type SetGrantsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*GrantResp  `json:"data"`
}

// WrapReq is the collection's key wrapped by the caller's app for a person, base64. Signature is the caller's
// Ed25519 key's, over what pkg/collectionkey says.
type WrapReq struct {
	UserID          string `json:"userId"`
	RecipientX25519 []byte `json:"recipientX25519" swaggertype:"string" format:"byte"`
	EphemeralPublic []byte `json:"ephemeralPublic" swaggertype:"string" format:"byte"`
	SealedKey       []byte `json:"sealedKey" swaggertype:"string" format:"byte"`
	Commitment      []byte `json:"commitment" swaggertype:"string" format:"byte"`
	Signature       []byte `json:"signature" swaggertype:"string" format:"byte"`
}

// WrapsReq is the key of a version wrapped for some people.
type WrapsReq struct {
	Version int        `json:"version"`
	Wraps   []*WrapReq `json:"wraps"`
}

func NewWrapsReq() *WrapsReq {
	return &WrapsReq{}
}

// Validate implements interface basedto.ReqValidator. Whether each wrap checks is the use case's to say.
func (req *WrapsReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateNumber(&req.Version, true, 1, math.MaxInt32, "version")
	validators = append(validators, basedto.ValidateSliceEx(req.Wraps, false, 1, wrapsMax, nil, "wraps")...)
	for _, w := range req.Wraps {
		if w == nil {
			validators = append(validators, basedto.ValidateCond(false, "wraps")...)
			continue
		}
		validators = append(validators, basedto.ValidateID(&w.UserID, true, "userId")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type WrapsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// WrapResp is the collection's key of a version, wrapped for the person, with who wrapped it: their app checks the
// signature with WrapperEd25519, and shows their fingerprint.
type WrapResp struct {
	Version            int    `json:"version"`
	RecipientX25519    []byte `json:"recipientX25519" swaggertype:"string" format:"byte"`
	EphemeralPublic    []byte `json:"ephemeralPublic" swaggertype:"string" format:"byte"`
	SealedKey          []byte `json:"sealedKey" swaggertype:"string" format:"byte"`
	Commitment         []byte `json:"commitment" swaggertype:"string" format:"byte"`
	Signature          []byte `json:"signature" swaggertype:"string" format:"byte"`
	WrappedBy          string `json:"wrappedBy"`
	WrapperUsername    string `json:"wrapperUsername"`
	WrapperX25519      []byte `json:"wrapperX25519" swaggertype:"string" format:"byte"`
	WrapperEd25519     []byte `json:"wrapperEd25519" swaggertype:"string" format:"byte"`
	WrapperFingerprint string `json:"wrapperFingerprint"`
}

// KeysResp is the collection's current key version, and the person's wraps, newest first.
type KeysResp struct {
	KeyVersion int         `json:"keyVersion"`
	Wraps      []*WrapResp `json:"wraps"`
}

type GetKeysResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *KeysResp     `json:"data"`
}
