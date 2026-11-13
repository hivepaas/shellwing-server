package orgdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
)

// OrgResp is an organization as one of its members sees it.
type OrgResp struct {
	ID                string       `json:"id"`
	Slug              string       `json:"slug"`
	Name              string       `json:"name"`
	Role              base.OrgRole `json:"role"`
	SSHCertTTLMinutes int          `json:"sshCertTtlMinutes"`
	// SSHPrincipals are what the member's certificates may log in as: their own accounts and their groups'.
	SSHPrincipals []string `json:"sshPrincipals"`
	// SSHCertExtensions is what a certificate the organization signs is allowed to do, by sshd's extension names.
	SSHCertExtensions []string `json:"sshCertExtensions"`
	// SSHKRLTokenCreatedAt is when the organization's servers were given a token to fetch the revocation list with.
	// Zero: there is none, and nothing fetches the list without a member's session. The token itself is never here.
	SSHKRLTokenCreatedAt time.Time `json:"sshKrlTokenCreatedAt,omitzero"`
}

// TransformOrg is the organization as one member sees it, with what their certificates may log in as.
func TransformOrg(org *entity.Organization, member *entity.OrgMember, principals []string) *OrgResp {
	if principals == nil {
		principals = []string{}
	}
	extensions := org.SSHCertExtensions
	if extensions == nil {
		extensions = []string{}
	}
	return &OrgResp{
		ID:                org.ID,
		Slug:              org.Slug,
		Name:              org.Name,
		Role:              member.Role,
		SSHCertTTLMinutes: org.SSHCertTTLMinutes,
		SSHPrincipals:     principals,

		SSHCertExtensions:    extensions,
		SSHKRLTokenCreatedAt: org.SSHKRLTokenCreatedAt,
	}
}

type ListOrgsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*OrgResp    `json:"data"`
}

// SSHCAResp is a CA's public side: its PublicKey is the line for sshd's TrustedUserCAKeys.
type SSHCAResp struct {
	ID        string           `json:"id"`
	Status    base.SSHCAStatus `json:"status"`
	PublicKey string           `json:"publicKey"`
	// KeyProvider is where the private half is: "local" for one this server holds, or the service holding it. The
	// key's name and version within that service are the operator's business rather than an admin's, so they are not
	// here.
	KeyProvider string    `json:"keyProvider"`
	CreatedAt   time.Time `json:"createdAt"`
	RetiredAt   time.Time `json:"retiredAt,omitzero"`
}

func TransformSSHCA(ca *entity.SSHCA) *SSHCAResp {
	provider := ca.KeyProvider
	if provider == "" {
		provider = base.KMSProviderLocal
	}
	return &SSHCAResp{
		ID:          ca.ID,
		Status:      ca.Status,
		PublicKey:   ca.PublicKey,
		KeyProvider: provider,
		CreatedAt:   ca.CreatedAt,
		RetiredAt:   ca.RetiredAt,
	}
}

type ListSSHCAsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*SSHCAResp  `json:"data"`
}

// SSHKRLTokenResp is a token a target server fetches the revocation list with. Token is filled in the answer that
// made it and never again: only its hash is kept.
type SSHKRLTokenResp struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateSSHKRLTokenResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *SSHKRLTokenResp `json:"data"`
}

type DeleteSSHKRLTokenResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type CreateSSHCAResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SSHCAResp    `json:"data"`
}

type DeleteSSHCAResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// UpdateSSHSettingsReq sets how long the organization's certificates last, and what they are allowed to do.
//
// Both fields are pointers, so that a caller changes one setting without having to send the other back. An empty list
// of extensions is a setting of its own - a certificate that logs in and runs a command - and would otherwise be
// indistinguishable from not having been sent; and a form that only changes the lifetime should not be able to write
// back a set of extensions it read some time ago.
type UpdateSSHSettingsReq struct {
	CertTTLMinutes *int `json:"certTtlMinutes"`
	// CertExtensions is what a certificate may do, from sshca.AllExtensions.
	CertExtensions *[]string `json:"certExtensions"`
}

func NewUpdateSSHSettingsReq() *UpdateSSHSettingsReq {
	return &UpdateSSHSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSHSettingsReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateNumber(req.CertTTLMinutes, false, base.SSHCertTTLMinMinutes,
		base.SSHCertTTLMaxMinutes, "certTtlMinutes")
	if req.CertExtensions != nil {
		validators = append(validators, basedto.ValidateSliceEx(*req.CertExtensions, true, 0,
			len(sshca.AllExtensions), sshca.AllExtensions, "certExtensions")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSHSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// MemberResp is a member of the organization. An admin sees all of it; a member sees who the others are and their
// role, and nothing of their email, status or accounts.
type MemberResp struct {
	UserID        string               `json:"userId"`
	Username      string               `json:"username"`
	FullName      string               `json:"fullName"`
	Role          base.OrgRole         `json:"role"`
	Email         string               `json:"email,omitzero"`
	Status        base.OrgMemberStatus `json:"status,omitzero"`
	AccountStatus base.UserStatus      `json:"accountStatus,omitzero"`
	SSHPrincipals []string             `json:"sshPrincipals,omitzero"`
	// Groups are the groups the member is in, whose SSH accounts are theirs too.
	Groups []*GroupRefResp `json:"groups,omitzero"`
}

func TransformMember(m *entity.OrgMemberUser, groups []*GroupRefResp) *MemberResp {
	principals := m.SSHPrincipals
	if principals == nil {
		principals = []string{}
	}
	if groups == nil {
		groups = []*GroupRefResp{}
	}
	return &MemberResp{
		Groups:        groups,
		UserID:        m.UserID,
		Username:      m.Username,
		FullName:      m.FullName,
		Role:          m.Role,
		Email:         m.Email,
		Status:        m.Status,
		AccountStatus: m.AccountStatus,
		SSHPrincipals: principals,
	}
}

// TransformMemberForMember is a member as another member sees them.
func TransformMemberForMember(m *entity.OrgMemberUser) *MemberResp {
	return &MemberResp{
		UserID:   m.UserID,
		Username: m.Username,
		FullName: m.FullName,
		Role:     m.Role,
	}
}

type ListMembersResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*MemberResp `json:"data"`
}

type SSHPrincipalsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []string      `json:"data"`
}

// SetSSHPrincipalsReq replaces the accounts a person's certificates may log in as. Empty takes them all away.
type SetSSHPrincipalsReq struct {
	Principals []string `json:"principals"`
}

func NewSetSSHPrincipalsReq() *SetSSHPrincipalsReq {
	return &SetSSHPrincipalsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetSSHPrincipalsReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateSliceEx(req.Principals, false, 0, sshca.MaxPrincipals, nil, "principals")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// IssueSSHCertificateReq carries the public key to certify. Its private half never leaves the device.
type IssueSSHCertificateReq struct {
	PublicKey string `json:"publicKey"`
}

func NewIssueSSHCertificateReq() *IssueSSHCertificateReq {
	return &IssueSSHCertificateReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *IssueSSHCertificateReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.PublicKey, true, 1, sshca.MaxPublicKeyLen, "publicKey")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// SSHCertificateResp is a certificate, as a line for a `-cert.pub` file, and what it says.
type SSHCertificateResp struct {
	Certificate string    `json:"certificate"`
	Principals  []string  `json:"principals"`
	ValidAfter  time.Time `json:"validAfter"`
	ValidBefore time.Time `json:"validBefore"`
}

type IssueSSHCertificateResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *SSHCertificateResp `json:"data"`
}

// UpdateMemberRoleReq makes a member an owner, an admin or a member.
type UpdateMemberRoleReq struct {
	Role base.OrgRole `json:"role"`
}

func NewUpdateMemberRoleReq() *UpdateMemberRoleReq {
	return &UpdateMemberRoleReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateMemberRoleReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateStrIn(&req.Role, true, base.AllOrgRoles, "role")...))
}

type UpdateMemberRoleResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// UpdateMemberStatusReq suspends a member, removes them, or makes them active again.
type UpdateMemberStatusReq struct {
	Status base.OrgMemberStatus `json:"status"`
}

func NewUpdateMemberStatusReq() *UpdateMemberStatusReq {
	return &UpdateMemberStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateMemberStatusReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateStrIn(&req.Status, true, base.AllOrgMemberStatuses, "status")...))
}

type UpdateMemberStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
