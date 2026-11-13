package datakeyserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/datakey"
	"github.com/hivepaas/shellwing-server/pkg/kms"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

// Load installs the data encryption key, creating one the first time the app runs.
//
// Opening the stored key needs whatever wraps it - the app secret, or a key
// management service this server can call - so a wrong secret, an unreachable
// service or a revoked credential fails here, at the start, rather than surfacing
// later as settings that cannot be read.
//
// If the configuration now names a different provider than the one that wrapped
// the stored key, the key is moved to it here. That is the whole of changing
// provider: one row, no pass over the values, and no window in which some of
// them are readable and some are not.
func (s *service) Load(ctx context.Context, db database.IDB) error {
	cfg := config.Current()
	if cfg == nil {
		return hperrors.NewMissing("Config")
	}
	configured, err := wrapperFor(ctx, cfg, cfg.KMS.Provider)
	if err != nil {
		return hperrors.Wrap(err)
	}

	stored, err := s.encryptionKeyRepo.GetActive(ctx, db)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if stored == nil {
		key, err := s.createKey(ctx, db, configured)
		if err != nil {
			return hperrors.Wrap(err)
		}
		datakey.SetActive(key)
		return nil
	}

	key, wrappedBy, err := s.openKey(ctx, cfg, stored.WrappedKey)
	if err != nil {
		return err
	}
	// Installed before the move: the key is the same either way, and a move that
	// fails should leave a server that works rather than one that cannot read
	// anything.
	datakey.SetActive(key)

	if wrappedBy != configured.Name() {
		if err = s.moveKey(ctx, db, key, wrappedBy, configured); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// openKey unwraps the stored key with the provider that wrapped it, and says
// which that was.
func (s *service) openKey(
	ctx context.Context,
	cfg *config.Config,
	stored string,
) (*datakey.Key, string, error) {
	provider := kms.ProviderOf(stored)
	wrapper, err := wrapperFor(ctx, cfg, provider)
	if err != nil {
		return nil, "", hperrors.Wrap(hperrors.ErrPreconditionFailed).WithCause(err).
			WithMsgLog("the stored data encryption key was wrapped by %s, which this server "+
				"is not configured to reach", provider)
	}
	raw, err := wrapper.Unwrap(ctx, stored)
	if err != nil {
		return nil, "", hperrors.Wrap(hperrors.ErrPreconditionFailed).WithCause(err).
			WithMsgLog("%s does not open the stored data encryption key", provider)
	}
	key, err := datakey.New(raw)
	if err != nil {
		return nil, "", hperrors.Wrap(hperrors.ErrPreconditionFailed).WithCause(err).
			WithMsgLog("%s opened the stored data encryption key into something that is not one", provider)
	}
	return key, provider, nil
}

// moveKey wraps the key with the configured provider and stores it that way.
//
// Only the wrapped form changes; the key itself does not, so nothing already
// encrypted is touched. A crash between wrapping and the update leaves the row as
// it was, which the old provider still opens, so the move is safe to retry on the
// next start.
func (s *service) moveKey(
	ctx context.Context,
	db database.IDB,
	key *datakey.Key,
	from string,
	to kms.Wrapper,
) error {
	wrapped, err := to.Wrap(ctx, key.Raw())
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		stored, err := s.encryptionKeyRepo.GetActive(ctx, db, bunex.SelectFor("UPDATE"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		if stored == nil {
			return hperrors.NewNotFound("Data encryption key")
		}
		// Another replica starting at the same time may have moved it already.
		// Both wrapped the same key, so either form is right; leave theirs.
		if kms.ProviderOf(stored.WrappedKey) == to.Name() {
			return nil
		}

		stored.WrappedKey = wrapped
		stored.UpdatedAt = timeutil.NowUTC()
		return hperrors.Wrap(s.encryptionKeyRepo.Update(ctx, db, stored,
			bunex.UpdateColumns("wrapped_key", "updated_at")))
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Worth a line in the log: it is a change to how the server's secrets are
	// protected, made because the configuration changed, and the operator should
	// be able to see that it took.
	s.logger.Warnf("the data encryption key is now wrapped by %s instead of %s", to.Name(), from)
	return nil
}

func (s *service) createKey(
	ctx context.Context,
	db database.IDB,
	wrapper kms.Wrapper,
) (*datakey.Key, error) {
	key, err := datakey.Generate()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	wrapped, err := wrapper.Wrap(ctx, key.Raw())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		// Another replica may have created one between the read above and here.
		// The unique index on is_active would reject the insert; take theirs.
		existing, err := s.encryptionKeyRepo.GetActive(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if existing != nil {
			key, _, err = s.openKey(ctx, config.Current(), existing.WrappedKey)
			return err
		}

		return s.encryptionKeyRepo.Insert(ctx, db, &entity.EncryptionKey{
			ID:         gofn.Must(ulid.NewStringULID()),
			WrappedKey: wrapped,
			IsActive:   true,
			CreatedAt:  timeNow,
			UpdatedAt:  timeNow,
		})
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return key, nil
}
