package datakeyserviceimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
	"github.com/hivepaas/shellwing-server/pkg/kms"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
)

// Rewrap re-encrypts the stored data key with a new app secret.
//
// This is the entire cost of changing the app secret: one row. The values in the
// settings table are encrypted with the data key, which does not change, so none
// of them is touched - no bulk pass, nothing to resume, no window in which some
// rows carry one key and some another.
//
// It is for the app secret only. When a key management service holds the key
// encryption key, the app secret does not wrap the data key and rotating it
// cannot be proved against it, so this refuses rather than reporting a rotation
// that protected nothing. Changing which service wraps the key is a change to the
// configuration, applied on the next start - see Load.
func (s *service) Rewrap(
	ctx context.Context,
	db database.IDB,
	currentAppSecret, newAppSecret string,
) error {
	current, err := kms.NewLocal(currentAppSecret)
	if err != nil {
		return hperrors.Wrap(err)
	}
	replacement, err := kms.NewLocal(newAppSecret)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return transaction.Execute(ctx, db, func(db database.Tx) error { //nolint:wrapcheck
		stored, err := s.encryptionKeyRepo.GetActive(ctx, db, bunex.SelectFor("UPDATE"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		if stored == nil {
			return hperrors.NewNotFound("Data encryption key")
		}
		if provider := kms.ProviderOf(stored.WrappedKey); provider != kms.ProviderLocal {
			return hperrors.Wrap(hperrors.ErrPreconditionFailed).
				WithMsgLog("the data encryption key is wrapped by %s, not by the app secret, "+
					"so changing the app secret does not re-wrap it", provider)
		}

		// Unwrapping with the current secret is what proves the caller gave the
		// right one: it either opens the key or it does not.
		raw, err := current.Unwrap(ctx, stored.WrappedKey)
		if err != nil {
			return hperrors.Wrap(hperrors.ErrForbidden).WithCause(err).
				WithMsgLog("the current app secret does not open the stored data encryption key")
		}
		// Built, rather than only unwrapped, so a secret that decrypts to the
		// wrong number of bytes is caught before it is stored back.
		if _, err = datakey.New(raw); err != nil {
			return hperrors.Wrap(err)
		}

		wrapped, err := replacement.Wrap(ctx, raw)
		if err != nil {
			return hperrors.Wrap(err)
		}

		stored.WrappedKey = wrapped
		stored.UpdatedAt = timeutil.NowUTC()
		if err := s.encryptionKeyRepo.Update(ctx, db, stored,
			bunex.UpdateColumns("wrapped_key", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}

		// The key itself is unchanged, so nothing in memory needs replacing.
		return nil
	})
}
