package config

import (
	"fmt"
	"strings"

	"github.com/hivepaas/shellwing-server/base"
)

// SSHCA says where a new certificate authority's private half is made.
//
// It is separate from KMS, which says what holds the key the stored secrets are
// under, because the two are different choices with different answers: a key
// that wraps and unwraps can be any symmetric key anywhere, while a key that
// signs an SSH certificate has to be an Ed25519 one, which AWS, GCP and Azure
// KMS cannot do. An installation can perfectly well keep its data encryption key
// in AWS KMS and its authorities in a Vault.
//
// It only applies to authorities made from now on. One already made keeps its
// key where that key is, because the key cannot be moved - a transit key is
// unexportable, and an authority's whole point is the public half the target
// servers were told to trust. Changing this and making a new authority is how an
// organization moves.
type SSHCA struct {
	// KeyProvider is one of base.SSHCAKeyProviders.
	KeyProvider string `toml:"key_provider" env:"SHW_SSH_CA_KEY_PROVIDER" default:"local"`
	// VaultKeyPrefix names the transit key of an authority: this, then the
	// authority's own id. One key per authority, so that retiring one and
	// revoking its certificates is a thing that can be done to it alone.
	VaultKeyPrefix string `toml:"vault_key_prefix" env:"SHW_SSH_CA_VAULT_KEY_PREFIX" default:"shellwing-ca-"`
}

// validate refuses a provider that cannot hold a signing key, at load rather
// than when somebody first tries to make an authority.
func (s *SSHCA) validate() error {
	if s.KeyProvider == "" {
		s.KeyProvider = base.KMSProviderLocal
	}
	if !base.ValidSSHCAKeyProvider(s.KeyProvider) {
		return fmt.Errorf("%w: SHW_SSH_CA_KEY_PROVIDER %q - one of %s",
			ErrSSHCAKeyProviderUnknown, s.KeyProvider, strings.Join(base.SSHCAKeyProviders, ", "))
	}
	if s.KeyProvider != base.KMSProviderLocal && s.VaultKeyPrefix == "" {
		s.VaultKeyPrefix = "shellwing-ca-"
	}
	return nil
}
