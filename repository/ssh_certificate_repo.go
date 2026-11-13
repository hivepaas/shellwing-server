package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// SSHCertificateRepo stores what each signed certificate says, and marks the ones that are refused from now on.
//
// Every Revoke call leaves alone a certificate that has ended, and one that was revoked already: sshd refuses an
// expired certificate on its own, and the first revocation is the one that is true. Each returns how many it marked,
// which the audit entry for the change that caused it records.
type SSHCertificateRepo interface {
	Insert(ctx context.Context, db database.IDB, cert *entity.SSHCertificate) error
	// RevokeOfUser refuses what one person holds in one organization, which is what a grant taken away, a membership
	// suspended or a membership ended comes down to.
	RevokeOfUser(ctx context.Context, db database.IDB, orgID, userID string, now time.Time) (int64, error)
	// RevokeOfUsers refuses what any of these people hold in one organization, for a change to a group they share.
	// No users is no work.
	RevokeOfUsers(ctx context.Context, db database.IDB, orgID string, userIDs []string, now time.Time) (int64, error)
	// RevokeAllOfUser refuses what one person holds in every organization, for an account disabled or deleted: the
	// account is what every membership hangs off, so none of them may still log in.
	RevokeAllOfUser(ctx context.Context, db database.IDB, userID string, now time.Time) (int64, error)
	// RevokeOfDevice refuses what one device holds, in every organization: a device revoked is revoked everywhere,
	// and its person may well still be a member.
	RevokeOfDevice(ctx context.Context, db database.IDB, deviceID string, now time.Time) (int64, error)
	// RevokeAllOfOrg refuses everything the organization has signed, for a change to what its certificates are
	// allowed to do: a certificate already held still allows what it did when it was signed.
	RevokeAllOfOrg(ctx context.Context, db database.IDB, orgID string, now time.Time) (int64, error)
	// RevokeOfCA refuses everything one authority signed, for an authority being replaced.
	RevokeOfCA(ctx context.Context, db database.IDB, orgID, caID string, now time.Time) (int64, error)
	// ListRevoked returns the organization's revoked certificates that have not ended yet - the ones a revocation
	// list has to name - the soonest to end first.
	ListRevoked(ctx context.Context, db database.IDB, orgID string, now time.Time) ([]*entity.SSHCertificate, error)
	// DeleteEnded clears out the certificates that ended before the given time, whether they were revoked or not.
	DeleteEnded(ctx context.Context, db database.IDB, before time.Time) (int64, error)
}

type sshCertificateRepo struct {
}

func NewSSHCertificateRepo() SSHCertificateRepo {
	return &sshCertificateRepo{}
}

func (repo *sshCertificateRepo) Insert(ctx context.Context, db database.IDB, cert *entity.SSHCertificate) error {
	if _, err := db.NewInsert().Model(cert).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sshCertificateRepo) RevokeOfUser(ctx context.Context, db database.IDB, orgID, userID string,
	now time.Time) (int64, error) {
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("org_id = ?", orgID).Where("user_id = ?", userID)
	})
}

func (repo *sshCertificateRepo) RevokeOfUsers(ctx context.Context, db database.IDB, orgID string, userIDs []string,
	now time.Time) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("org_id = ?", orgID).Where("user_id IN (?)", bun.List(userIDs))
	})
}

func (repo *sshCertificateRepo) RevokeAllOfUser(ctx context.Context, db database.IDB, userID string,
	now time.Time) (int64, error) {
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("user_id = ?", userID)
	})
}

func (repo *sshCertificateRepo) RevokeOfDevice(ctx context.Context, db database.IDB, deviceID string,
	now time.Time) (int64, error) {
	if deviceID == "" {
		return 0, nil
	}
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("device_id = ?", deviceID)
	})
}

func (repo *sshCertificateRepo) RevokeAllOfOrg(ctx context.Context, db database.IDB, orgID string,
	now time.Time) (int64, error) {
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("org_id = ?", orgID)
	})
}

func (repo *sshCertificateRepo) RevokeOfCA(ctx context.Context, db database.IDB, orgID, caID string,
	now time.Time) (int64, error) {
	return repo.revoke(ctx, db, now, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Where("org_id = ?", orgID).Where("ca_id = ?", caID)
	})
}

// revoke marks what narrow picks out, among the certificates still worth marking.
func (repo *sshCertificateRepo) revoke(ctx context.Context, db database.IDB, now time.Time,
	narrow func(q *bun.UpdateQuery) *bun.UpdateQuery) (int64, error) {
	q := narrow(db.NewUpdate().Model((*entity.SSHCertificate)(nil)).
		Set("revoked_at = ?", now).
		Where("revoked_at IS NULL").
		Where("valid_before > ?", now))
	result, err := q.Exec(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	marked, err := result.RowsAffected()
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return marked, nil
}

func (repo *sshCertificateRepo) ListRevoked(ctx context.Context, db database.IDB, orgID string,
	now time.Time) ([]*entity.SSHCertificate, error) {
	var certs []*entity.SSHCertificate
	err := db.NewSelect().Model(&certs).
		Where("org_id = ?", orgID).
		Where("revoked_at IS NOT NULL").
		Where("valid_before > ?", now).
		Order("valid_before ASC", "serial ASC").Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return certs, nil
}

func (repo *sshCertificateRepo) DeleteEnded(ctx context.Context, db database.IDB, before time.Time) (int64, error) {
	result, err := db.NewDelete().Model((*entity.SSHCertificate)(nil)).
		Where("valid_before <= ?", before).Exec(ctx)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return deleted, nil
}
