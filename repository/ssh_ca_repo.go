package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
)

// SSHCARepo stores certificate authorities. Every call names the organization: one's CA is never reached through
// another's.
type SSHCARepo interface {
	Insert(ctx context.Context, db database.IDB, ca *entity.SSHCA) error
	// GetActive returns nil when the organization has no active CA.
	GetActive(ctx context.Context, db database.IDB, orgID string) (*entity.SSHCA, error)
	// Get returns nil when the organization has no such CA.
	Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.SSHCA, error)
	// List returns the organization's CAs, the newest first.
	List(ctx context.Context, db database.IDB, orgID string) ([]*entity.SSHCA, error)
	// ListTrusted returns the CAs a target server trusts: the one signing, and any published beside it. Newest first.
	ListTrusted(ctx context.Context, db database.IDB, orgID string) ([]*entity.SSHCA, error)
	// RetireActive ends the active CA, if any, and destroys its private key.
	RetireActive(ctx context.Context, db database.IDB, orgID string, now time.Time) error
	// SetStatus moves one CA to a status. Retiring it destroys its private key and records when.
	SetStatus(ctx context.Context, db database.IDB, orgID, id string, status base.SSHCAStatus,
		now time.Time) error
	// DemoteActive makes the active CA published, so that another may take over signing. It keeps its private key:
	// a published authority has not stopped being one, it has only stopped signing.
	DemoteActive(ctx context.Context, db database.IDB, orgID string) error
	Delete(ctx context.Context, db database.IDB, orgID, id string) error
}

type sshCARepo struct {
}

func NewSSHCARepo() SSHCARepo {
	return &sshCARepo{}
}

func (repo *sshCARepo) Insert(ctx context.Context, db database.IDB, ca *entity.SSHCA) error {
	if _, err := db.NewInsert().Model(ca).Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sshCARepo) GetActive(ctx context.Context, db database.IDB, orgID string) (*entity.SSHCA, error) {
	ca := &entity.SSHCA{}
	err := db.NewSelect().Model(ca).
		Where("org_id = ?", orgID).Where("status = ?", base.SSHCAStatusActive).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return ca, nil
}

func (repo *sshCARepo) Get(ctx context.Context, db database.IDB, orgID, id string) (*entity.SSHCA, error) {
	ca := &entity.SSHCA{}
	err := db.NewSelect().Model(ca).Where("org_id = ?", orgID).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return ca, nil
}

func (repo *sshCARepo) List(ctx context.Context, db database.IDB, orgID string) ([]*entity.SSHCA, error) {
	var cas []*entity.SSHCA
	err := db.NewSelect().Model(&cas).Where("org_id = ?", orgID).Order("created_at DESC", "id DESC").Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return cas, nil
}

func (repo *sshCARepo) RetireActive(ctx context.Context, db database.IDB, orgID string, now time.Time) error {
	_, err := db.NewUpdate().Model((*entity.SSHCA)(nil)).
		Set("status = ?", base.SSHCAStatusRetired).
		Set("private_key = NULL").
		Set("retired_at = ?", now).
		Where("org_id = ?", orgID).Where("status = ?", base.SSHCAStatusActive).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sshCARepo) ListTrusted(ctx context.Context, db database.IDB, orgID string) (
	[]*entity.SSHCA, error) {
	cas := []*entity.SSHCA{}
	err := db.NewSelect().Model(&cas).
		Where("org_id = ?", orgID).
		Where("status IN (?)", bun.List([]base.SSHCAStatus{base.SSHCAStatusActive, base.SSHCAStatusPublished})).
		Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return cas, nil
}

func (repo *sshCARepo) SetStatus(ctx context.Context, db database.IDB, orgID, id string,
	status base.SSHCAStatus, now time.Time) error {
	query := db.NewUpdate().Model((*entity.SSHCA)(nil)).
		Set("status = ?", status).
		Where("org_id = ?", orgID).Where("id = ?", id)
	if status == base.SSHCAStatusRetired {
		// The key goes with the status: a retired authority is one nothing can be signed with again, and the column
		// is what makes that true rather than a rule somebody has to remember.
		query = query.Set("private_key = NULL").Set("retired_at = ?", now)
	}
	if _, err := query.Exec(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sshCARepo) DemoteActive(ctx context.Context, db database.IDB, orgID string) error {
	_, err := db.NewUpdate().Model((*entity.SSHCA)(nil)).
		Set("status = ?", base.SSHCAStatusPublished).
		Where("org_id = ?", orgID).Where("status = ?", base.SSHCAStatusActive).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sshCARepo) Delete(ctx context.Context, db database.IDB, orgID, id string) error {
	_, err := db.NewDelete().Model((*entity.SSHCA)(nil)).Where("org_id = ?", orgID).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
