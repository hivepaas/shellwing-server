package config

import (
	"fmt"
	"strings"

	"github.com/hivepaas/shellwing-server/base"
)

// KMS says what holds the key encryption key: the key that wraps the data
// encryption key every stored secret is sealed with.
//
// The default is the app secret, which needs nothing deployed beside the server.
// Naming a service instead means the database and its backups are no longer
// enough to read what is in them - see pkg/kms for what that does and does not
// buy.
//
// The fields are flat rather than nested per provider because the loader reads
// one level of section, and because an environment variable per field is what a
// container is configured with.
type KMS struct {
	// Provider is one of base.KMSProviders.
	Provider string `toml:"provider" env:"SHW_KMS_PROVIDER" default:"local"`

	// VaultAddress is the base URL of a HashiCorp Vault or an OpenBao, such as
	// https://vault.internal:8200.
	VaultAddress string `toml:"vault_address" env:"SHW_KMS_VAULT_ADDRESS"`
	// VaultToken authenticates this server to it.
	VaultToken string `toml:"vault_token" env:"SHW_KMS_VAULT_TOKEN"`
	// VaultMount is where the transit engine is mounted.
	VaultMount string `toml:"vault_mount" env:"SHW_KMS_VAULT_MOUNT" default:"transit"`
	// VaultKey is the transit key's name.
	VaultKey string `toml:"vault_key" env:"SHW_KMS_VAULT_KEY"`
	// VaultNamespace is Vault Enterprise's namespace, if it is in one.
	VaultNamespace string `toml:"vault_namespace" env:"SHW_KMS_VAULT_NAMESPACE"`
	// VaultCACertFile is a PEM file to trust instead of the system store, for the
	// private authority a self-hosted Vault is usually behind.
	VaultCACertFile string `toml:"vault_ca_cert_file" env:"SHW_KMS_VAULT_CA_CERT_FILE"`

	// AWSKeyID names the KMS key: its ID, its ARN, or an alias such as
	// alias/shellwing-dek.
	AWSKeyID string `toml:"aws_key_id" env:"SHW_KMS_AWS_KEY_ID"`
	// AWSRegion is the key's region. Empty, the ambient AWS configuration decides.
	AWSRegion string `toml:"aws_region" env:"SHW_KMS_AWS_REGION"`
	// AWSAccessKeyID and AWSSecretAccessKey are an explicit credential. Leaving
	// both empty uses the SDK's own chain, which picks up an instance role or a
	// service account's - the better way, since there is then no long-lived
	// secret in this configuration at all.
	AWSAccessKeyID     string `toml:"aws_access_key_id" env:"SHW_KMS_AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `toml:"aws_secret_access_key" env:"SHW_KMS_AWS_SECRET_ACCESS_KEY"`
	AWSSessionToken    string `toml:"aws_session_token" env:"SHW_KMS_AWS_SESSION_TOKEN"`
	// AWSEndpointURL overrides where the service is, for an AWS-compatible one or
	// a local test double.
	AWSEndpointURL string `toml:"aws_endpoint_url" env:"SHW_KMS_AWS_ENDPOINT_URL"`
}

// validate refuses a provider this server does not know, at load rather than at
// the first wrap. A typo would otherwise be found only once the server had
// started, by which point it has already decided it cannot read its own secrets.
//
// What each provider needs is not checked here: that belongs where the wrapper is
// built, which is the one place that knows, and it reports the field by name.
func (k *KMS) validate() error {
	if k.Provider == "" {
		// The loader's default did not apply, which happens when the section is
		// present in a config file with the key left out.
		k.Provider = base.KMSProviderLocal
		return nil
	}
	if !base.ValidKMSProvider(k.Provider) {
		return fmt.Errorf("%w: SHW_KMS_PROVIDER %q - one of %s",
			ErrKMSProviderUnknown, k.Provider, strings.Join(base.KMSProviders, ", "))
	}
	return nil
}
