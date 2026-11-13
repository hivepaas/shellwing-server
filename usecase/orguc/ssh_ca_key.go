package orguc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"
	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/kms"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

// Where an authority's private half lives, and how it signs wherever that is.
//
// Two places so far. A local authority has its key here, as a PEM sealed with the server's data key, which is what
// every authority made before this was. One in a transit engine has no key here at all: the engine signs, and will not
// hand the key over - so taking this server, its database and its backups is no longer enough to sign a certificate
// that every host in the fleet accepts. That is the one thing moving the key buys, and it is the reason the seam is
// here rather than inside pkg/sshca, which knows nothing about services.
//
// Which one a new authority gets is the operator's choice, in config.SSHCA, and applies only to authorities made from
// then on: a key in an engine cannot be moved out, and a local one moved in would be a new public half and so a new
// authority anyway.

// caKeySettings is where a new authority's key is made, and what to name it there.
//
// config.Current is nil until the app loads one, which it does before anything serves; a process without one is a
// test, and the answer there is the default every installation starts with.
func caKeySettings() *config.SSHCA {
	if cfg := config.Current(); cfg != nil {
		return &cfg.SSHCA
	}
	return &config.SSHCA{KeyProvider: base.KMSProviderLocal}
}

// vaultOptionsFor is how to reach the engine for one authority's key. The address, token and mount are the
// installation's, shared with whatever wraps the data encryption key; only the key name is the authority's own.
func vaultOptionsFor(keyName string) kms.VaultTransitOptions {
	cfg := config.Current()
	if cfg == nil {
		return kms.VaultTransitOptions{Key: keyName}
	}
	return kms.VaultTransitOptions{
		Address:    cfg.KMS.VaultAddress,
		Token:      cfg.KMS.VaultToken,
		Mount:      cfg.KMS.VaultMount,
		Key:        keyName,
		Namespace:  cfg.KMS.VaultNamespace,
		CACertFile: cfg.KMS.VaultCACertFile,
	}
}

// caSigner is the authority as something that signs.
//
// For a key in an engine this needs no call: the public half is already written down beside the authority, as the line
// the target servers were told to trust, so one signature costs one call rather than two.
func (uc *UC) caSigner(ctx context.Context, ca *entity.SSHCA) (ssh.Signer, error) {
	if ca.KeyIsLocal() {
		privateKey, err := ca.GetPrivateKey()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		signer, err := sshca.SignerFromPEM(privateKey)
		return signer, hperrors.Wrap(err)
	}
	if ca.KeyProvider != base.KMSProviderVaultTransit {
		return nil, hperrors.Wrap(hperrors.ErrSSHCAKeyUnreachable).WithParam("Name", ca.KeyProvider)
	}

	public, err := sshca.Ed25519PublicKeyOf(ca.PublicKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	signer, err := kms.NewVaultTransitSigner(ctx, vaultOptionsFor(ca.KeyName), &kms.SigningKey{
		Name:      ca.KeyName,
		Version:   ca.KeyVersion,
		PublicKey: public,
	})
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHCAKeyUnreachable).WithCause(err).
			WithParam("Name", ca.KeyProvider)
	}
	sshSigner, err := ssh.NewSignerFromSigner(signer)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return sshSigner, nil
}

// newCA makes an authority's key where the configuration says, and the row that names it.
//
// The key is made before anything is stored, and for an engine the signer is built from what the engine answered: an
// authority is never recorded unless it can actually sign, because one that cannot is an organization that can issue
// no certificates at all until an admin notices and replaces it.
func (uc *UC) newCA(ctx context.Context, orgID string, now time.Time) (*entity.SSHCA, error) {
	ca := &entity.SSHCA{
		ID:        gofn.Must(ulid.NewStringULID()),
		OrgID:     orgID,
		Status:    base.SSHCAStatusActive,
		CreatedAt: now,
	}

	settings := caKeySettings()
	provider := settings.KeyProvider
	if provider == "" || provider == base.KMSProviderLocal {
		key, err := sshca.GenerateCA()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		ca.KeyProvider = base.KMSProviderLocal
		ca.PublicKey = key.PublicKey
		if err = ca.SetPrivateKey(key.PrivateKey); err != nil {
			return nil, hperrors.Wrap(err)
		}
		return ca, nil
	}

	// One key per authority, named after it: retiring one and revoking what it signed is then something that can be
	// done to it alone, and an operator reading the engine's key list can tell which authority each key belongs to.
	opts := vaultOptionsFor(settings.VaultKeyPrefix + ca.ID)
	signing, err := kms.CreateVaultTransitSigningKey(ctx, opts)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHCAKeyUnreachable).WithCause(err).WithParam("Name", provider)
	}
	signer, err := kms.NewVaultTransitSigner(ctx, opts, signing)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHCAKeyUnreachable).WithCause(err).WithParam("Name", provider)
	}
	sshSigner, err := ssh.NewSignerFromSigner(signer)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	ca.KeyProvider = base.KMSProviderVaultTransit
	ca.KeyName = signing.Name
	ca.KeyVersion = signing.Version
	// The engine's own answer rather than anything derived here: it is what will be trusted, so it should come from
	// the thing that will do the signing.
	ca.PublicKey = sshca.PublicKeyLine(sshSigner)
	return ca, nil
}
