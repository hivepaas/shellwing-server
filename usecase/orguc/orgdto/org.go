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
	SSHPrincipals     []string     `json:"sshPrincipals"`
}

func TransformOrg(org *entity.Organization, member *entity.OrgMember) *OrgResp {
	principals := member.SSHPrincipals
	if principals == nil {
		principals = []string{}
	}
	return &OrgResp{
		ID:                org.ID,
		Slug:              org.Slug,
		Name:              org.Name,
		Role:              member.Role,
		SSHCertTTLMinutes: org.SSHCertTTLMinutes,
		SSHPrincipals:     principals,
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
	CreatedAt time.Time        `json:"createdAt"`
	RetiredAt time.Time        `json:"retiredAt,omitzero"`
}

func TransformSSHCA(ca *entity.SSHCA) *SSHCAResp {
	return &SSHCAResp{
		ID:        ca.ID,
		Status:    ca.Status,
		PublicKey: ca.PublicKey,
		CreatedAt: ca.CreatedAt,
		RetiredAt: ca.RetiredAt,
	}
}

type ListSSHCAsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*SSHCAResp  `json:"data"`
}

type CreateSSHCAResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SSHCAResp    `json:"data"`
}

type DeleteSSHCAResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// UpdateSSHSettingsReq sets how long the organization's certificates last.
type UpdateSSHSettingsReq struct {
	CertTTLMinutes int `json:"certTtlMinutes"`
}

func NewUpdateSSHSettingsReq() *UpdateSSHSettingsReq {
	return &UpdateSSHSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSHSettingsReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateNumber(&req.CertTTLMinutes, true, base.SSHCertTTLMinMinutes,
		base.SSHCertTTLMaxMinutes, "certTtlMinutes")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSHSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// MemberResp is a member of the organization as an admin sees them: who they are, and what they may log in as.
type MemberResp struct {
	UserID        string               `json:"userId"`
	Username      string               `json:"username"`
	Email         string               `json:"email"`
	FullName      string               `json:"fullName"`
	Role          base.OrgRole         `json:"role"`
	Status        base.OrgMemberStatus `json:"status"`
	SSHPrincipals []string             `json:"sshPrincipals"`
}

func TransformMember(m *entity.OrgMemberUser) *MemberResp {
	principals := m.SSHPrincipals
	if principals == nil {
		principals = []string{}
	}
	return &MemberResp{
		UserID:        m.UserID,
		Username:      m.Username,
		Email:         m.Email,
		FullName:      m.FullName,
		Role:          m.Role,
		Status:        m.Status,
		SSHPrincipals: principals,
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
