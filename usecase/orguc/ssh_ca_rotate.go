package orguc

import (
	"context"
	"strings"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

// Rotating an authority, in the two steps it has to take.
//
// A new authority cannot start signing the moment it is made. Its public key has to reach every target server first,
// or everybody whose server has not been updated is refused - which is why a new one is published rather than made
// active, and why promoting it is a separate thing an admin does once the servers have it.
//
// The one it replaces stays published afterwards rather than being retired, so the certificates it already signed keep
// working until they expire. Retiring it is the third step, and the only one that refuses anything.
//
// Replacing an authority outright is still here, for the case rotating is not for: the key is compromised, and
// nothing it signed should be taken again. That is CreateSSHCAReq.Replace, and it is what this used to always do.

// SSHCAStatus moves an authority between signing, published and retired.
func (uc *UC) SSHCAStatus(ctx context.Context, auth *basedto.Auth, orgID, caID string,
	req *orgdto.SetSSHCAStatusReq) (*orgdto.SSHCAStatusResp, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var data *orgdto.SSHCAResp
	var revoked int64
	now := timeutil.NowUTC()
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		if err := uc.requireAdmin(ctx, tx, auth, orgID); err != nil {
			return err
		}
		// Two admins rotating at once would otherwise both demote and both promote.
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
		if ca.Status != base.SSHCAStatusPublished {
			// Only a published authority moves: the one signing has to be replaced rather than demoted on its own,
			// or an organization would be left with nothing that signs, and a retired one does not come back.
			return hperrors.Wrap(hperrors.ErrSSHCANotPublished).WithParam("Name", string(ca.Status))
		}

		if revoked, err = uc.applySSHCAStatus(ctx, tx, ca, req.Status, now); err != nil {
			return err
		}
		ca.Status = req.Status
		data = orgdto.TransformSSHCA(ca)

		section := auditSectionCAPromote
		if req.Status == base.SSHCAStatusRetired {
			section = auditSectionCARetire
		}
		return uc.recordCAChange(ctx, tx, auth, orgID, section, ca.ID,
			auditdetail.New().Set("publicKey", ca.PublicKey).Set("certsRevoked", revoked))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &orgdto.SSHCAStatusResp{Data: data}, nil
}

// applySSHCAStatus is the change itself, once the authority is known to be a published one.
func (uc *UC) applySSHCAStatus(ctx context.Context, tx database.Tx, ca *entity.SSHCA,
	status base.SSHCAStatus, now time.Time) (int64, error) {
	if status == base.SSHCAStatusActive {
		// Demoted first: exactly one authority signs, and the index says so.
		if err := uc.sshCARepo.DemoteActive(ctx, tx, ca.OrgID); err != nil {
			return 0, hperrors.Wrap(err)
		}
		if err := uc.sshCARepo.SetStatus(ctx, tx, ca.OrgID, ca.ID, status, now); err != nil {
			return 0, hperrors.Wrap(err)
		}
		// Nothing is revoked: what the old authority signed is still good on every server that trusts it, which is
		// every server, because it stays published.
		return 0, nil
	}

	if err := uc.sshCARepo.SetStatus(ctx, tx, ca.OrgID, ca.ID, status, now); err != nil {
		return 0, hperrors.Wrap(err)
	}
	// Retiring is the one step that refuses something. A server may still trust the key until it fetches the list
	// again, so what this authority signed goes into the revocation list rather than being left to expire.
	revoked, err := uc.sshCertRepo.RevokeOfCA(ctx, tx, ca.OrgID, ca.ID, now)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return revoked, nil
}

// writeSSHCAKeys is the TrustedUserCAKeys file: one line per authority a target server should trust.
//
// Served rather than copied by hand because that is what makes rotating work. A new authority is published before it
// signs, and a server that fetches this on a timer has it before anything is signed with it - which is the whole
// difference between rotating an authority and locking everybody out.
func (uc *UC) writeSSHCAKeys(ctx context.Context, orgID string) ([]byte, error) {
	cas, err := uc.sshCARepo.ListTrusted(ctx, uc.db, orgID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	lines := make([]string, 0, len(cas))
	for _, ca := range cas {
		lines = append(lines, strings.TrimSpace(ca.PublicKey))
	}
	if len(lines) == 0 {
		// An organization with no authority yet. An empty file is still an answer: a server told to read it and
		// unable to refuses every certificate, which is worse than one that trusts nothing.
		return []byte{}, nil
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}
