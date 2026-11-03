package licenseserviceimpl

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/licensesig"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/licenseservice"
)

const (
	// The portal takes an installation ID of 25 lowercase letters and digits.
	installationIDLen      = 25
	installationIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func New(licenseRepo repository.LicenseRepo) licenseservice.Service {
	return &service{licenseRepo: licenseRepo}
}

type service struct {
	licenseRepo repository.LicenseRepo
}

func newInstallationID() string {
	data := make([]byte, installationIDLen)
	_, _ = rand.Read(data) // never fails: see crypto/rand
	for i, b := range data {
		// A little bias does no harm: this names a server, and hides nothing.
		data[i] = installationIDAlphabet[int(b)%len(installationIDAlphabet)]
	}
	return string(data)
}

func (s *service) EnsureInstallation(ctx context.Context, db database.IDB) error {
	state := &entity.LicenseState{ID: 1, InstallationID: newInstallationID(), UpdatedAt: timeutil.NowUTC()}
	return hperrors.Wrap(s.licenseRepo.InsertIfAbsent(ctx, db, state))
}

func (s *service) verifier() (*licensesig.Verifier, error) {
	verifier, err := licensesig.Embedded(config.Current().DevMode.Enabled)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return verifier, nil
}

// freeUserLimit is the limit with no license in force. A server in development may raise it, to work on more users
// than that: a seed has more.
func freeUserLimit() int {
	cfg := config.Current()
	if cfg.DevMode.Enabled && cfg.DevMode.UserLimit > 0 {
		return cfg.DevMode.UserLimit
	}
	return base.FreeUserLimit
}

// evaluate works out what is in force from the key at the time given.
func evaluate(installationID string, key *licensesig.Payload, verifyErr error, hasKey bool,
	now time.Time) *licenseservice.State {
	state := &licenseservice.State{
		Status:         base.LicenseStatusNone,
		InstallationID: installationID,
		MaxUsers:       freeUserLimit(),
		FreeUsers:      freeUserLimit(),
	}
	if !hasKey {
		return state
	}
	if verifyErr != nil || key == nil {
		state.Status = base.LicenseStatusInvalid
		return state
	}
	state.Key = key
	if !key.ForInstallation(installationID) {
		state.Status = base.LicenseStatusOtherInstallation
		return state
	}
	state.GraceEndsAt = key.GraceEndsAt()
	switch {
	case now.Before(key.ExpiresAt) && key.Trial:
		state.Status = base.LicenseStatusTrial
	case now.Before(key.ExpiresAt):
		state.Status = base.LicenseStatusValid
	case now.Before(state.GraceEndsAt):
		state.Status = base.LicenseStatusGrace
	default:
		state.Status = base.LicenseStatusExpired
		return state
	}
	state.Plan = key.Plan
	state.MaxUsers = base.NoUserLimit
	if limit, limited := key.Limits[licensesig.LimitUsers]; limited {
		state.MaxUsers = limit
	}
	return state
}

func (s *service) load(ctx context.Context, db database.IDB) (*entity.LicenseState, error) {
	row, err := s.licenseRepo.Get(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if row == nil {
		return nil, hperrors.NewNotFound("License state")
	}
	return row, nil
}

func (s *service) Current(ctx context.Context, db database.IDB) (*licenseservice.State, error) {
	row, err := s.load(ctx, db)
	if err != nil {
		return nil, err
	}
	return s.stateOf(row)
}

func (s *service) stateOf(row *entity.LicenseState) (*licenseservice.State, error) {
	if row.KeyText == "" {
		return evaluate(row.InstallationID, nil, nil, false, timeutil.NowUTC()), nil
	}
	verifier, err := s.verifier()
	if err != nil {
		return nil, err
	}
	key, verifyErr := verifier.Verify(row.KeyText)
	return evaluate(row.InstallationID, key, verifyErr, true, timeutil.NowUTC()), nil
}

func (s *service) CountUserSeats(ctx context.Context, db database.IDB) (int, error) {
	count, err := s.licenseRepo.CountUserSeats(ctx, db)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return count, nil
}

func (s *service) CheckUserSeat(ctx context.Context, tx database.Tx) error {
	if err := s.licenseRepo.LockUserSeats(ctx, tx); err != nil {
		return hperrors.Wrap(err)
	}
	state, err := s.Current(ctx, tx)
	if err != nil {
		return err
	}
	if state.MaxUsers == base.NoUserLimit {
		return nil
	}
	count, err := s.licenseRepo.CountUserSeats(ctx, tx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if count >= state.MaxUsers {
		return hperrors.Wrap(hperrors.ErrPlanLimitReached).WithParam("Max", state.MaxUsers).
			WithMsgLog("%d users hold a seat, of %d", count, state.MaxUsers)
	}
	return nil
}

func (s *service) Install(ctx context.Context, db database.IDB, keyText string) (*licenseservice.State, error) {
	row, err := s.load(ctx, db)
	if err != nil {
		return nil, err
	}
	keyText = licensesig.NormalizeKeyText(keyText)
	verifier, err := s.verifier()
	if err != nil {
		return nil, err
	}
	key, err := verifier.Verify(keyText)
	if err != nil {
		if errors.Is(err, licensesig.ErrKeyInvalid) || errors.Is(err, licensesig.ErrInvalidPayload) {
			// The cause stays out of the answer: it can hold nothing the caller needs.
			return nil, hperrors.Wrap(hperrors.ErrLicenseKeyInvalid)
		}
		return nil, hperrors.Wrap(err)
	}
	state := evaluate(row.InstallationID, key, nil, true, timeutil.NowUTC())
	switch state.Status { //nolint:exhaustive // the others put a key in force
	case base.LicenseStatusOtherInstallation:
		return nil, hperrors.Wrap(hperrors.ErrLicenseKeyOtherInstallation).
			WithParam("InstallationID", row.InstallationID)
	case base.LicenseStatusExpired:
		return nil, hperrors.Wrap(hperrors.ErrLicenseKeyExpired)
	}
	row.KeyText = keyText
	row.UpdatedAt = timeutil.NowUTC()
	if err = s.licenseRepo.Update(ctx, db, row); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return state, nil
}

func (s *service) Remove(ctx context.Context, db database.IDB) (*licenseservice.State, error) {
	row, err := s.load(ctx, db)
	if err != nil {
		return nil, err
	}
	row.KeyText = ""
	row.UpdatedAt = timeutil.NowUTC()
	if err = s.licenseRepo.Update(ctx, db, row); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return s.stateOf(row)
}
