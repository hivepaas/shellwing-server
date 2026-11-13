package orguc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/hivepaas/shellwing-server/pkg/reqinfo"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
	"github.com/hivepaas/shellwing-server/pkg/sshkrl"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/repository/cacherepository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

const (
	auditSectionCACreate   = "ca-create"
	auditSectionCADelete   = "ca-delete"
	auditSectionSettings   = "settings"
	auditSectionPrincipals = "principals"
	auditSectionKRLToken   = "krl-token"
	auditSectionRole       = "role"
	auditSectionStatus     = "status"
)

// How many certificates may be signed in an hour, for one person and for one of their devices. A device asks for one
// when it has none that still lasts, so a handful a day is the shape of ordinary use; what these bound is a session or
// a device token being used to stock up. The limits are per actor rather than per address, for the reason
// pkg/failbackoff gives: the caller is already authenticated and can change address freely.
const (
	certIssueWindow     = time.Hour
	certIssuesPerUser   = 60
	certIssuesPerDevice = 30
)

// The revocation list is fetched without a session, so what one address may cost is bounded here. A server fetching
// every minute uses 60 of these an hour; the limit leaves room for several behind one address.
const (
	krlFetchWindow  = time.Hour
	krlFetchesPerIP = 2000
	// krlTokenLen is the length in bytes of a fetch token, before hex encoding doubles it.
	krlTokenLen = 32
)

// endedCertRetention is how long the record of a certificate is kept after it has ended. A revocation list has no use
// for one - sshd refuses an expired certificate on its own - so what the time buys is the audit trail's other half:
// the audit log says a certificate was signed, and this says what was in it.
const endedCertRetention = 7 * 24 * time.Hour

type UC struct {
	db             *database.DB
	orgRepo        repository.OrganizationRepo
	sshCARepo      repository.SSHCARepo
	sshCertRepo    repository.SSHCertificateRepo
	groupRepo      repository.OrgGroupRepo
	collectionRepo repository.CollectionRepo
	rateLimitRepo  cacherepository.RateLimitRepo
	auditService   auditservice.Service
}

func New(db *database.DB, orgRepo repository.OrganizationRepo, sshCARepo repository.SSHCARepo,
	sshCertRepo repository.SSHCertificateRepo, groupRepo repository.OrgGroupRepo,
	collectionRepo repository.CollectionRepo, rateLimitRepo cacherepository.RateLimitRepo,
	auditService auditservice.Service) *UC {
	return &UC{db: db, orgRepo: orgRepo, sshCARepo: sshCARepo, sshCertRepo: sshCertRepo, groupRepo: groupRepo,
		collectionRepo: collectionRepo, rateLimitRepo: rateLimitRepo, auditService: auditService}
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
	if !isAdmin(member) {
		return hperrors.NewForbidden("Organization")
	}
	return nil
}

func isAdmin(member *entity.OrgMember) bool {
	return member.Role == base.OrgRoleOwner || member.Role == base.OrgRoleAdmin
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
		if org == nil {
			continue
		}
		principals, err := uc.principalsOf(ctx, uc.db, member)
		if err != nil {
			return nil, err
		}
		data = append(data, orgdto.TransformOrg(org, member, principals))
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

// SSHKRL writes the organization's revocation list, the file a target server's RevokedKeys names.
//
// It is for any member, as the CA's public key is: both belong to a server's configuration, and what this says beyond
// the public key is a set of serial numbers, each drawn at random. A list with nothing in it is still a list: a server
// told to read the file and unable to refuses every certificate, so there is always something to fetch.
//
// Every authority the organization has had is covered, not only the one signing now. A certificate the retired one
// signed is still good on a server that still trusts it.
func (uc *UC) SSHKRL(ctx context.Context, auth *basedto.Auth, orgID string) ([]byte, error) {
	if _, _, err := uc.memberOf(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	return uc.writeSSHKRL(ctx, orgID)
}

// SSHKRLForToken writes the list for a target server fetching it with the organization's token, which is all the
// credential it has and all it needs. A token no organization has is simply not found, and says no more than that.
//
// The lookup is by the token's hash, so an equal hash is an equal token and there is nothing to compare in
// non-constant time. A caller guessing is bounded by the limit below rather than by the odds, which are already
// hopeless for a token of this length.
func (uc *UC) SSHKRLForToken(ctx context.Context, orgID, token string) ([]byte, error) {
	if err := uc.limitKRLFetches(ctx); err != nil {
		return nil, err
	}
	if token == "" {
		return nil, hperrors.Wrap(hperrors.ErrSSHKRLTokenInvalid)
	}
	org, err := uc.orgRepo.GetBySSHKRLToken(ctx, uc.db, hashKRLToken(token))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The organization asked for has to be the token's own. The token alone would be enough to find it, but then a
	// server configured with one organization's address and another's token would quietly fetch the wrong list.
	if org == nil || org.ID != orgID || org.Status != base.OrgStatusActive {
		return nil, hperrors.Wrap(hperrors.ErrSSHKRLTokenInvalid)
	}
	return uc.writeSSHKRL(ctx, org.ID)
}

// CreateSSHKRLToken makes the organization a new token for its servers to fetch the list with, and returns it. It is
// returned here and nowhere else: only its hash is kept. One before it stops working, so a token that may have leaked
// is replaced by making another.
func (uc *UC) CreateSSHKRLToken(ctx context.Context, auth *basedto.Auth, orgID string) (
	*orgdto.CreateSSHKRLTokenResp, error) {
	token := gofn.RandTokenAsHex(krlTokenLen)
	now := timeutil.NowUTC()
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		if err := uc.orgRepo.SetSSHKRLToken(ctx, tx, orgID, hashKRLToken(token), now); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionKRLToken, orgID,
			auditdetail.New().Set("created", true))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.CreateSSHKRLTokenResp{Data: &orgdto.SSHKRLTokenResp{Token: token, CreatedAt: now}}, nil
}

// DeleteSSHKRLToken takes the token away. The servers holding it stop being able to fetch, and what they have in
// place keeps working: an old list refuses less than the current one, never more.
func (uc *UC) DeleteSSHKRLToken(ctx context.Context, auth *basedto.Auth, orgID string) (
	*orgdto.DeleteSSHKRLTokenResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		if err := uc.orgRepo.SetSSHKRLToken(ctx, tx, orgID, "", timeutil.NowUTC()); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionKRLToken, orgID,
			auditdetail.New().Set("created", false))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.DeleteSSHKRLTokenResp{}, nil
}

// limitKRLFetches bounds what an unauthenticated caller costs. The limit is by address, not by actor, because there is
// no actor until a token has been recognized; a server fetching every minute is far inside it, and a proxy in front
// of several of them is still inside it.
func (uc *UC) limitKRLFetches(ctx context.Context) error {
	clientIP := gofn.Coalesce(reqinfo.ClientIPFrom(ctx), "unknown")
	count, err := uc.rateLimitRepo.Hit(ctx, "ssh-krl-fetch:ip:"+clientIP, krlFetchWindow)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count > krlFetchesPerIP {
		return hperrors.Wrap(hperrors.ErrTooManyRequests).
			WithMsgLog("revocation list fetches from %s over %d an hour", clientIP, krlFetchesPerIP)
	}
	return nil
}

// hashKRLToken is what is stored and what is looked up, as a device's refresh token is kept.
func hashKRLToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (uc *UC) writeSSHKRL(ctx context.Context, orgID string) ([]byte, error) {
	now := timeutil.NowUTC()
	// Certificates that ended a while ago are cleared out here, which is where the table is read anyway and is a
	// path the target servers themselves keep warm. It belongs in a scheduled job, and should move to one once the
	// worker runs recurring work; until then this keeps a row per certificate ever signed from piling up, and the
	// delay leaves them readable for as long as anybody is likely to be looking into a login that happened.
	if _, err := uc.sshCertRepo.DeleteEnded(ctx, uc.db, now.Add(-endedCertRetention)); err != nil {
		return nil, hperrors.Wrap(err)
	}
	cas, err := uc.sshCARepo.List(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	certs, err := uc.sshCertRepo.ListRevoked(ctx, uc.db, orgID, now)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	serialsOf := make(map[string][]uint64, len(cas))
	for _, cert := range certs {
		serialsOf[cert.CAID] = append(serialsOf[cert.CAID], uint64(cert.Serial)) //nolint:gosec // G115: within 63 bits
	}
	authorities := make([]sshkrl.Authority, 0, len(cas))
	for _, ca := range cas {
		authorities = append(authorities, sshkrl.Authority{PublicKey: ca.PublicKey, Serials: serialsOf[ca.ID]})
	}
	// The time of writing is the version: it is both meaningful and always ahead of the list written before, which is
	// what `ssh-keygen -k -u` compares when somebody merges this into a list of their own.
	krl, err := sshkrl.Write(uint64(now.Unix()), now, authorities) //nolint:gosec // G115: a time after 1970
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return krl, nil
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
		retiring, err := uc.sshCARepo.GetActive(ctx, tx, orgID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err := uc.sshCARepo.RetireActive(ctx, tx, orgID, now); err != nil {
			return hperrors.Wrap(err)
		}
		// What the retired authority signed is refused from now on. A server that still trusts its public key would
		// otherwise keep taking those certificates until each ended, which is the opposite of what replacing an
		// authority is for; the app asks for a new certificate when it next needs one.
		var revoked int64
		if retiring != nil {
			if revoked, err = uc.sshCertRepo.RevokeOfCA(ctx, tx, orgID, retiring.ID, now); err != nil {
				return hperrors.Wrap(err)
			}
		}
		if err := uc.sshCARepo.Insert(ctx, tx, ca); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionCACreate, ca.ID,
			auditdetail.New().Set("publicKey", ca.PublicKey).Set("certsRevoked", revoked))
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

// UpdateSSHSettings sets how long certificates last from now on, and what they are allowed to do.
//
// A new lifetime reaches only the certificates signed after it: one already signed has its end written into it, and
// shortening the setting cannot shorten that. A change to what certificates may do is different - it revokes the ones
// already held, since each of those still allows what it allowed when it was signed, and the app asks for another.
func (uc *UC) UpdateSSHSettings(ctx context.Context, auth *basedto.Auth, orgID string,
	req *orgdto.UpdateSSHSettingsReq) (*orgdto.UpdateSSHSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		org, _, err := uc.memberOf(ctx, tx, auth, orgID)
		if err != nil {
			return err
		}
		if err = uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		detail := auditdetail.New()
		if req.CertTTLMinutes != nil {
			if err = uc.orgRepo.SetSSHCertTTL(ctx, tx, orgID, *req.CertTTLMinutes); err != nil {
				return hperrors.Wrap(err)
			}
			detail = detail.Set("certTtlMinutes", *req.CertTTLMinutes)
		}
		if req.CertExtensions != nil {
			extensions := cleanExtensions(*req.CertExtensions)
			var revoked int64
			if !slices.Equal(orEmpty(org.SSHCertExtensions), extensions) {
				if err = uc.orgRepo.SetSSHCertExtensions(ctx, tx, orgID, extensions); err != nil {
					return hperrors.Wrap(err)
				}
				if revoked, err = uc.sshCertRepo.RevokeAllOfOrg(ctx, tx, orgID, timeutil.NowUTC()); err != nil {
					return hperrors.Wrap(err)
				}
			}
			detail = detail.Set("certExtensions", extensions).Set("certsRevoked", revoked)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionSettings, orgID, detail)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.UpdateSSHSettingsResp{}, nil
}

// ListMembers lists the organization's members. An admin sees everything of each, removed members included; a member
// sees the active ones, by name and role only.
func (uc *UC) ListMembers(ctx context.Context, auth *basedto.Auth, orgID string) (*orgdto.ListMembersResp, error) {
	_, me, err := uc.memberOf(ctx, uc.db, auth, orgID)
	if err != nil {
		return nil, err
	}
	members, err := uc.orgRepo.ListMembers(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	groupsOf := map[string][]*orgdto.GroupRefResp{}
	if isAdmin(me) {
		inGroups, err := uc.groupRepo.ListMembers(ctx, uc.db, orgID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, m := range inGroups {
			groupsOf[m.UserID] = append(groupsOf[m.UserID], &orgdto.GroupRefResp{ID: m.GroupID, Name: m.GroupName})
		}
	}
	data := make([]*orgdto.MemberResp, 0, len(members))
	for _, m := range members {
		switch {
		case isAdmin(me):
			data = append(data, orgdto.TransformMember(m, groupsOf[m.UserID]))
		case m.Status == base.OrgMemberStatusActive && m.AccountStatus == base.UserStatusActive:
			data = append(data, orgdto.TransformMemberForMember(m))
		}
	}
	return &orgdto.ListMembersResp{Data: data}, nil
}

// UpdateMemberRole makes a member an owner, an admin or a member. Owners and admins change members and admins; only an
// owner makes an owner or changes one, and the last active owner stays one.
func (uc *UC) UpdateMemberRole(ctx context.Context, auth *basedto.Auth, orgID, userID string,
	req *orgdto.UpdateMemberRoleReq) (*orgdto.UpdateMemberRoleResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		me, member, err := uc.lockMember(ctx, tx, auth, orgID, userID)
		if err != nil {
			return err
		}
		if !isAdmin(me) {
			return hperrors.NewForbidden("Organization")
		}
		if member.Status == base.OrgMemberStatusRemoved {
			return hperrors.NewNotFound("Member")
		}
		if (member.Role == base.OrgRoleOwner || req.Role == base.OrgRoleOwner) && me.Role != base.OrgRoleOwner {
			return hperrors.NewForbidden("Organization")
		}
		if member.Role == req.Role {
			return nil
		}
		if member.Role == base.OrgRoleOwner {
			if err = uc.requireAnotherOwner(ctx, tx, orgID, userID); err != nil {
				return err
			}
		}
		from := member.Role
		member.Role = req.Role
		if err = uc.orgRepo.UpdateMember(ctx, tx, member); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordMemberChange(ctx, tx, auth, orgID, userID, auditSectionRole, string(from), string(req.Role))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.UpdateMemberRoleResp{}, nil
}

// UpdateMemberStatus suspends a member, removes them, or makes them active again. Owners and admins do it to members
// and admins, and only an owner to an owner; anybody may remove themselves, which is leaving. The last active owner
// stays one. A removed member loses their role, accounts, groups and grants: brought back, they are a member with none.
func (uc *UC) UpdateMemberStatus(ctx context.Context, auth *basedto.Auth, orgID, userID string,
	req *orgdto.UpdateMemberStatusReq) (*orgdto.UpdateMemberStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		me, member, err := uc.lockMember(ctx, tx, auth, orgID, userID)
		if err != nil {
			return err
		}
		if err = mayChangeStatus(auth, me, member, req.Status); err != nil {
			return err
		}
		if member.Status == req.Status {
			return nil
		}
		if member.Role == base.OrgRoleOwner && member.Status == base.OrgMemberStatusActive {
			if err = uc.requireAnotherOwner(ctx, tx, orgID, userID); err != nil {
				return err
			}
		}
		from := member.Status
		member.Status = req.Status
		if req.Status == base.OrgMemberStatusRemoved {
			if err = uc.strip(ctx, tx, member); err != nil {
				return err
			}
		}
		if err = uc.orgRepo.UpdateMember(ctx, tx, member); err != nil {
			return hperrors.Wrap(err)
		}
		// Somebody stopped, for a while or for good, keeps no certificate: the organization gives them nothing from
		// here on, and what they hold would otherwise work until it ended.
		if req.Status != base.OrgMemberStatusActive {
			if _, err = uc.sshCertRepo.RevokeOfUser(ctx, tx, orgID, userID, timeutil.NowUTC()); err != nil {
				return hperrors.Wrap(err)
			}
		}
		return uc.recordMemberChange(ctx, tx, auth, orgID, userID, auditSectionStatus, string(from),
			string(req.Status))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.UpdateMemberStatusResp{}, nil
}

// mayChangeStatus says whether the person acting may set this member's status. Anybody may remove themselves - that
// is leaving - and otherwise it takes an admin, and an owner to touch an owner.
func mayChangeStatus(auth *basedto.Auth, me, member *entity.OrgMember, to base.OrgMemberStatus) error {
	if member.UserID == auth.UserID() && to == base.OrgMemberStatusRemoved {
		return nil
	}
	if !isAdmin(me) {
		return hperrors.NewForbidden("Organization")
	}
	if member.Role == base.OrgRoleOwner && me.Role != base.OrgRoleOwner {
		return hperrors.NewForbidden("Organization")
	}
	return nil
}

// strip takes from a member being removed their role, accounts, groups and grants.
func (uc *UC) strip(ctx context.Context, tx database.Tx, member *entity.OrgMember) error {
	member.Role = base.OrgRoleMember
	member.SSHPrincipals = []string{}
	if err := uc.groupRepo.RemoveFromAll(ctx, tx, member.OrgID, member.UserID); err != nil {
		return hperrors.Wrap(err)
	}
	if err := uc.collectionRepo.RemoveUserGrants(ctx, tx, member.OrgID, member.UserID); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// lockMember returns the person acting and the member they act on, with the organization locked: two changes at once
// cannot both take away what each thought the other left, such as the last owner.
func (uc *UC) lockMember(ctx context.Context, tx database.Tx, auth *basedto.Auth, orgID, userID string) (
	me, member *entity.OrgMember, err error) {
	if _, err = uc.orgRepo.LockByID(ctx, tx, orgID); err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if _, me, err = uc.memberOf(ctx, tx, auth, orgID); err != nil {
		return nil, nil, err
	}
	if member, err = uc.orgRepo.GetMember(ctx, tx, orgID, userID); err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if member == nil {
		return nil, nil, hperrors.NewNotFound("Member")
	}
	return me, member, nil
}

func (uc *UC) requireAnotherOwner(ctx context.Context, db database.IDB, orgID, userID string) error {
	owners, err := uc.orgRepo.CountActiveOwners(ctx, db, orgID, userID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if owners == 0 {
		return hperrors.Wrap(hperrors.ErrOrgLastOwner)
	}
	return nil
}

func (uc *UC) recordMemberChange(ctx context.Context, db database.IDB, auth *basedto.Auth, orgID, userID, section,
	from, to string) error {
	return uc.record(ctx, db, auth, &auditservice.Entry{
		Type:    base.AuditLogTypeOrgMemberUpdate,
		Section: section,
		ResType: base.ResourceTypeOrgMember,
		ResID:   userID,
		Detail:  auditdetail.New().Set("orgId", orgID).Set("from", from).Set("to", to).String(),
	})
}

// GetSSHPrincipals returns the accounts a member's certificates may log in as.
func (uc *UC) GetSSHPrincipals(ctx context.Context, auth *basedto.Auth, orgID, userID string) (
	*orgdto.SSHPrincipalsResp, error) {
	if err := uc.requireAdmin(ctx, uc.db, auth, orgID); err != nil {
		return nil, err
	}
	member, err := uc.orgRepo.GetMember(ctx, uc.db, orgID, userID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if member == nil || member.Status == base.OrgMemberStatusRemoved {
		return nil, hperrors.NewNotFound("Member")
	}
	return &orgdto.SSHPrincipalsResp{Data: orEmpty(member.SSHPrincipals)}, nil
}

// SetSSHPrincipals replaces what a member's certificates may log in as, from the next one on. Principals only ever come
// from here: a person's username is not one, or whoever could take the name "root" would get root on every server that
// trusts the CA. An admin cannot grant themselves - see noSelfGrant.
func (uc *UC) SetSSHPrincipals(ctx context.Context, auth *basedto.Auth, orgID, userID string,
	req *orgdto.SetSSHPrincipalsReq) (*orgdto.SSHPrincipalsResp, error) {
	principals, err := cleanPrincipals(req.Principals)
	if err != nil {
		return nil, err
	}
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		member, err := uc.orgRepo.GetMember(ctx, tx, orgID, userID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if member == nil || member.Status == base.OrgMemberStatusRemoved {
			return hperrors.NewNotFound("Member")
		}
		err = uc.noSelfGrant(ctx, tx, auth, orgID, func() error {
			return hperrors.Wrap(uc.orgRepo.SetSSHPrincipals(ctx, tx, orgID, userID, principals))
		})
		if err != nil {
			return err
		}
		revoked, err := uc.sshCertRepo.RevokeOfUser(ctx, tx, orgID, userID, timeutil.NowUTC())
		if err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, auditSectionPrincipals, userID,
			auditdetail.New().Set("principals", principals).Set("certsRevoked", revoked))
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
	if err = uc.limitCertIssues(ctx, auth); err != nil {
		return nil, err
	}
	var data *orgdto.SSHCertificateResp
	err = transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		org, member, err := uc.memberOf(ctx, tx, auth, orgID)
		if err != nil {
			return err
		}
		principals, err := uc.principalsOf(ctx, tx, member)
		if err != nil {
			return err
		}
		if len(principals) == 0 {
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
			Principals: principals,
			Extensions: org.SSHCertExtensions,
			TTL:        time.Duration(org.SSHCertTTLMinutes) * time.Minute,
			Now:        timeutil.NowUTC(),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		// Recorded in the same transaction as the audit entry, and for the same reason: what is handed out is what
		// can be named in a revocation list afterwards.
		if err = uc.recordCert(ctx, tx, auth, orgID, ca.ID, principals, signed); err != nil {
			return err
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
				Set("principals", principals).
				Set("validBefore", signed.ValidBefore).String(),
		})
		if err != nil {
			return err
		}
		data = &orgdto.SSHCertificateResp{
			Certificate: signed.Certificate,
			Principals:  principals,
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

// limitCertIssues counts a request for a certificate against the person, and against the device if an app is asking,
// and refuses it once either is over its limit. The device is counted first, so one device already over its own limit
// no longer uses up the person's: their other devices keep working.
//
// A request that goes on to be refused has still been counted. That is deliberate: a caller guessing at an
// organization it is not a member of, or asking with no accounts granted, is the shape this is here to bound.
func (uc *UC) limitCertIssues(ctx context.Context, auth *basedto.Auth) error {
	if device := deviceIDOf(auth); device != "" {
		count, err := uc.rateLimitRepo.Hit(ctx, "ssh-cert-issue:device:"+device, certIssueWindow)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if count > certIssuesPerDevice {
			return hperrors.Wrap(hperrors.ErrTooManyRequests).
				WithMsgLog("certificates for device %s over %d an hour", device, certIssuesPerDevice)
		}
	}
	count, err := uc.rateLimitRepo.Hit(ctx, "ssh-cert-issue:user:"+auth.UserID(), certIssueWindow)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count > certIssuesPerUser {
		return hperrors.Wrap(hperrors.ErrTooManyRequests).
			WithMsgLog("certificates for user %s over %d an hour", auth.UserID(), certIssuesPerUser)
	}
	return nil
}

// recordCert keeps what the certificate says, which is what lets it be named in a revocation list afterwards. It is
// written in the transaction that hands the certificate out, for the reason the audit entry beside it is.
func (uc *UC) recordCert(ctx context.Context, tx database.Tx, auth *basedto.Auth, orgID, caID string,
	principals []string, signed *sshca.Signed) error {
	return hperrors.Wrap(uc.sshCertRepo.Insert(ctx, tx, &entity.SSHCertificate{
		Serial:      int64(signed.Serial), //nolint:gosec // G115: pkg/sshca draws one within 63 bits
		OrgID:       orgID,
		CAID:        caID,
		UserID:      auth.UserID(),
		DeviceID:    deviceIDOf(auth),
		Fingerprint: signed.Fingerprint,
		KeyID:       keyIDOf(auth),
		Principals:  principals,
		ValidAfter:  signed.ValidAfter,
		ValidBefore: signed.ValidBefore,
	}))
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

// principalsOf is what a member's certificates may log in as: their own accounts and their groups'.
func (uc *UC) principalsOf(ctx context.Context, db database.IDB, member *entity.OrgMember) ([]string, error) {
	fromGroups, err := uc.groupRepo.PrincipalsOf(ctx, db, member.OrgID, member.UserID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	principals := append(slices.Clone(member.SSHPrincipals), fromGroups...)
	slices.Sort(principals)
	return orEmpty(slices.Compact(principals)), nil
}

// noSelfGrant runs a change that may alter who may log in as what, and refuses it if it would give the admin making
// it an account they did not already have.
//
// It compares what the person's own certificates would name before and after rather than checking each path's
// arguments, because there are several ways to the same grant: their own accounts, a group's accounts, and who is in
// a group. A check per path would have to be right in all three, and in whatever path comes next.
//
// Taking access away from themselves is allowed, and so is granting anybody else. What is refused is one admin
// account being enough to reach a server: another admin grants it, and the audit log says who did.
func (uc *UC) noSelfGrant(ctx context.Context, tx database.Tx, auth *basedto.Auth, orgID string,
	change func() error) error {
	before, err := uc.effectivePrincipals(ctx, tx, orgID, auth.UserID())
	if err != nil {
		return err
	}
	if err = change(); err != nil {
		return err
	}
	after, err := uc.effectivePrincipals(ctx, tx, orgID, auth.UserID())
	if err != nil {
		return err
	}
	for _, name := range after {
		if !slices.Contains(before, name) {
			return hperrors.Wrap(hperrors.ErrSSHSelfGrant).WithParam("Name", name)
		}
	}
	return nil
}

// effectivePrincipals is what the person's certificates would name as things stand: their own accounts and their
// groups'. Nothing at all for somebody who is no longer a member, since they get no certificates.
func (uc *UC) effectivePrincipals(ctx context.Context, db database.IDB, orgID, userID string) ([]string, error) {
	member, err := uc.orgRepo.GetMember(ctx, db, orgID, userID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if member == nil || member.Status == base.OrgMemberStatusRemoved {
		return nil, nil
	}
	return uc.principalsOf(ctx, db, member)
}

// cleanExtensions keeps the extensions in the order sshca.AllExtensions lists them, once each, so that two ways of
// asking for the same thing are stored the same way and a change can be told from a reorder.
func cleanExtensions(names []string) []string {
	extensions := make([]string, 0, len(sshca.AllExtensions))
	for _, name := range sshca.AllExtensions {
		if slices.Contains(names, name) {
			extensions = append(extensions, name)
		}
	}
	return extensions
}

// cleanPrincipals sorts the names, keeps each once, and refuses one no account could have.
func cleanPrincipals(names []string) ([]string, error) {
	principals := slices.Clone(names)
	slices.Sort(principals)
	principals = slices.Compact(principals)
	for _, name := range principals {
		if !sshca.ValidPrincipal(name) {
			return nil, hperrors.Wrap(hperrors.ErrSSHPrincipalInvalid).WithParam("Name", name)
		}
	}
	return orEmpty(principals), nil
}

func orEmpty(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}
