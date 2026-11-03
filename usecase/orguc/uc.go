package orguc

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

const (
	auditSectionCACreate   = "ca-create"
	auditSectionCADelete   = "ca-delete"
	auditSectionSettings   = "settings"
	auditSectionPrincipals = "principals"
)

type UC struct {
	db           *database.DB
	orgRepo      repository.OrganizationRepo
	sshCARepo    repository.SSHCARepo
	auditService auditservice.Service
}

func New(db *database.DB, orgRepo repository.OrganizationRepo, sshCARepo repository.SSHCARepo,
	auditService auditservice.Service) *UC {
	return &UC{db: db, orgRepo: orgRepo, sshCARepo: sshCARepo, auditService: auditService}
}

// memberOf is the person's membership of an active organization. An organization they are not an active member of is
// simply not found: saying more would tell a caller which ids exist.
func (uc *UC) memberOf(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID string) (
	*entity.Organization, *entity.OrgMember, error) {
	org, err := uc.orgRepo.GetByID(ctx, db, orgID)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if org == nil || org.Status != base.OrgStatusActive {
		return nil, nil, hperrors.NewNotFound("Organization")
	}
	member, err := uc.orgRepo.GetMember(ctx, db, orgID, auth.UserID())
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if member == nil || member.Status != base.OrgMemberStatusActive {
		return nil, nil, hperrors.NewNotFound("Organization")
	}
	return org, member, nil
}

func (uc *UC) requireAdmin(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID string) error {
	_, member, err := uc.memberOf(ctx, db, auth, orgID)
	if err != nil {
		return err
	}
	if member.Role != base.OrgRoleOwner && member.Role != base.OrgRoleAdmin {
		return hperrors.NewForbidden("Organization")
	}
	return nil
}

func (uc *UC) record(ctx context.Context, db database.IDB, auth *basedto.Auth, entry *auditservice.Entry) error {
	entry.Scope = base.ObjectScopeGlobal
	entry.Source = base.AuditLogSourceAPIUpdate
	entry.Auth = auth
	return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, entry))
}

func (uc *UC) recordCAChange(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID, section, resID string,
	detail *auditdetail.Builder) error {
	return uc.record(ctx, db, auth, &auditservice.Entry{
		Type:    base.AuditLogTypeSSHCAUpdate,
		Section: section,
		ResType: base.ResourceTypeSSHCA,
		ResID:   resID,
		Detail:  detail.Set("orgId", orgID).String(),
	})
}

// ListOrgs returns the organizations the person is an active member of.
func (uc *UC) ListOrgs(ctx context.Context, auth *basedto.Auth) (*orgdto.ListOrgsResp, error) {
	members, err := uc.orgRepo.ListMembershipsOfUser(ctx, uc.db, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := make([]*orgdto.OrgResp, 0, len(members))
	for _, member := range members {
		org, err := uc.orgRepo.GetByID(ctx, uc.db, member.OrgID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if org != nil {
			data = append(data, orgdto.TransformOrg(org, member))
		}
	}
	return &orgdto.ListOrgsResp{Data: data}, nil
}

// ListSSHCAs returns the organization's CAs. Their public keys are for any member: a person setting up a server needs
// them.
func (uc *UC) ListSSHCAs(ctx context.Context, auth *basedto.Auth, orgID string) (*orgdto.ListSSHCAsResp, error) {
	if _, _, err := uc.memberOf(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	cas, err := uc.sshCARepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := make([]*orgdto.SSHCAResp, 0, len(cas))
	for _, ca := range cas {
		data = append(data, orgdto.TransformSSHCA(ca))
	}
	return &orgdto.ListSSHCAsResp{Data: data}, nil
}

// CreateSSHCA makes a new active CA. One before it is retired and its private key destroyed: certificates it signed
// still work until they expire, on servers that still trust it.
func (uc *UC) CreateSSHCA(ctx context.Context, auth *basedto.Auth, orgID string) (*orgdto.CreateSSHCAResp, error) {
	key, err := sshca.GenerateCA()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	now := timeutil.NowUTC()
	ca := &entity.SSHCA{
		ID:        gofn.Must(ulid.NewStringULID()),
		OrgID:     orgID,
		Status:    base.SSHCAStatusActive,
		PublicKey: key.PublicKey,
		CreatedAt: now,
	}
	if err = ca.SetPrivateKey(key.PrivateKey); err != nil {
		return nil, hperrors.Wrap(err)
	}
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		// Two at once: the second waits, and retires the first.
		if _, err := uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
			return hperrors.Wrap(err)
		}
		if err := uc.sshCARepo.RetireActive(ctx, tx, orgID, now); err != nil {
			return hperrors.Wrap(err)
		}
		if err := uc.sshCARepo.Insert(ctx, tx, ca); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionCACreate, ca.ID,
			auditdetail.New().Set("publicKey", ca.PublicKey))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.CreateSSHCAResp{Data: orgdto.TransformSSHCA(ca)}, nil
}

// DeleteSSHCA removes a retired CA from the list. The active one signs, and is replaced rather than deleted.
func (uc *UC) DeleteSSHCA(ctx context.Context, auth *basedto.Auth, orgID, caID string) (
	*orgdto.DeleteSSHCAResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		if _, err := uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
			return hperrors.Wrap(err)
		}
		ca, err := uc.sshCARepo.Get(ctx, tx, orgID, caID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if ca == nil {
			return hperrors.NewNotFound("SSH certificate authority")
		}
		if ca.Status == base.SSHCAStatusActive {
			return hperrors.Wrap(hperrors.ErrSSHCAActive)
		}
		if err = uc.sshCARepo.Delete(ctx, tx, orgID, caID); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionCADelete, caID,
			auditdetail.New().Set("publicKey", ca.PublicKey))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.DeleteSSHCAResp{}, nil
}

// UpdateSSHSettings sets how long certificates last from now on. Those signed already keep their end.
func (uc *UC) UpdateSSHSettings(ctx context.Context, auth *basedto.Auth, orgID string,
	req *orgdto.UpdateSSHSettingsReq) (*orgdto.UpdateSSHSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		if err := uc.orgRepo.SetSSHCertTTL(ctx, tx, orgID, req.CertTTLMinutes); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionSettings, orgID,
			auditdetail.New().Set("certTtlMinutes", req.CertTTLMinutes))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.UpdateSSHSettingsResp{}, nil
}

// GetSSHPrincipals returns the accounts a member's certificates may log in as.
// ListMembers lists the organization's members, for an admin granting them accounts.
func (uc *UC) ListMembers(ctx context.Context, auth *basedto.Auth, orgID string) (*orgdto.ListMembersResp, error) {
	if err := uc.requireAdmin(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	members, err := uc.orgRepo.ListMembers(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := make([]*orgdto.MemberResp, 0, len(members))
	for _, m := range members {
		data = append(data, orgdto.TransformMember(m))
	}
	return &orgdto.ListMembersResp{Data: data}, nil
}

func (uc *UC) GetSSHPrincipals(ctx context.Context, auth *basedto.Auth, orgID, userID string) (
	*orgdto.SSHPrincipalsResp, error) {
	if err := uc.requireAdmin(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	member, err := uc.orgRepo.GetMember(ctx, uc.db, orgID, userID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if member == nil {
		return nil, hperrors.NewNotFound("Member")
	}
	return &orgdto.SSHPrincipalsResp{Data: orEmpty(member.SSHPrincipals)}, nil
}

// SetSSHPrincipals replaces what a member's certificates may log in as, from the next one on. Principals only ever come
// from here: a person's username is not one, or whoever could take the name "root" would get root on every server that
// trusts the CA.
func (uc *UC) SetSSHPrincipals(ctx context.Context, auth *basedto.Auth, orgID, userID string,
	req *orgdto.SetSSHPrincipalsReq) (*orgdto.SSHPrincipalsResp, error) {
	principals := slices.Clone(req.Principals)
	slices.Sort(principals)
	principals = slices.Compact(principals)
	for _, name := range principals {
		if !sshca.ValidPrincipal(name) {
			return nil, hperrors.Wrap(hperrors.ErrSSHPrincipalInvalid).WithParam("Name", name)
		}
	}
	principals = orEmpty(principals)
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		member, err := uc.orgRepo.GetMember(ctx, tx, orgID, userID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if member == nil {
			return hperrors.NewNotFound("Member")
		}
		if err = uc.orgRepo.SetSSHPrincipals(ctx, tx, orgID, userID, principals); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionPrincipals, userID,
			auditdetail.New().Set("principals", principals))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.SSHPrincipalsResp{Data: principals}, nil
}

// IssueSSHCertificate signs the person's public key. What the certificate may do comes from the organization, never
// from the request. It is recorded in the same transaction: a certificate nobody has a record of is not handed out.
func (uc *UC) IssueSSHCertificate(ctx context.Context, auth *basedto.Auth, orgID string,
	req *orgdto.IssueSSHCertificateReq) (*orgdto.IssueSSHCertificateResp, error) {
	key, err := sshca.ParseUserKey(req.PublicKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var data *orgdto.SSHCertificateResp
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		org, member, err := uc.memberOf(ctx, tx, auth, orgID)
		if err != nil {
			return err
		}
		if len(member.SSHPrincipals) == 0 {
			return hperrors.Wrap(hperrors.ErrSSHNoPrincipals)
		}
		ca, err := uc.sshCARepo.GetActive(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if ca == nil {
			return hperrors.Wrap(hperrors.ErrSSHCANotSetUp)
		}
		caKey, err := ca.GetPrivateKey()
		if err != nil {
			return hperrors.Wrap(err)
		}
		signed, err := sshca.Sign(caKey, &sshca.UserCert{
			Key:        key,
			KeyID:      keyIDOf(auth),
			Principals: member.SSHPrincipals,
			TTL:        time.Duration(org.SSHCertTTLMinutes) * time.Minute,
			Now:        timeutil.NowUTC(),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = uc.record(ctx, tx, auth, &auditservice.Entry{
			Type:    base.AuditLogTypeSSHCertIssue,
			ResType: base.ResourceTypeSSHCert,
			ResID:   strconv.FormatUint(signed.Serial, 10),
			Detail: auditdetail.New().
				Set("orgId", orgID).
				Set("caId", ca.ID).
				Set("deviceId", deviceIDOf(auth)).
				Set("fingerprint", signed.Fingerprint).
				Set("principals", member.SSHPrincipals).
				Set("validBefore", signed.ValidBefore).String(),
		})
		if err != nil {
			return err
		}
		data = &orgdto.SSHCertificateResp{
			Certificate: signed.Certificate,
			Principals:  member.SSHPrincipals,
			ValidAfter:  signed.ValidAfter,
			ValidBefore: signed.ValidBefore,
		}
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.IssueSSHCertificateResp{Data: data}, nil
}

// deviceIDOf is the app that asks: an app's token has the device as its session. Empty for a person on the web app.
func deviceIDOf(auth *basedto.Auth) string {
	claims := auth.User.AuthClaims
	if claims == nil || claims.Scope != jwtsession.ScopeSync {
		return ""
	}
	return claims.UID
}

// keyIDOf is what sshd logs for a login with the certificate, so that a server's own logs say who it was.
func keyIDOf(auth *basedto.Auth) string {
	id := fmt.Sprintf("user=%s id=%s", auth.User.Username, auth.UserID())
	if device := deviceIDOf(auth); device != "" {
		id += " device=" + device
	}
	return id
}

func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}
