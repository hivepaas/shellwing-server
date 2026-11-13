package base

// The key management services that can hold the key encryption key: the key
// that wraps the data encryption key every stored secret is sealed with. See
// pkg/kms for what each one is and what choosing it buys.
//
// They are here, in the leaf package, rather than in pkg/kms where the code is,
// because config has to validate the name it was given and cannot import
// pkg/kms: that package reaches hperrors, which reaches logging, which reads
// config. pkg/kms re-exports these under shorter names.
const (
	// KMSProviderLocal wraps with the app secret from the configuration.
	KMSProviderLocal = "local"
	// KMSProviderVaultTransit wraps with the transit engine of HashiCorp Vault
	// or OpenBao.
	KMSProviderVaultTransit = "vault-transit"
	// KMSProviderAWSKMS wraps with a symmetric key in AWS KMS.
	KMSProviderAWSKMS = "awskms"
)

// KMSProviders are the names a configuration may give, in the order a message
// listing them should read.
var KMSProviders = []string{KMSProviderLocal, KMSProviderVaultTransit, KMSProviderAWSKMS}

// ValidKMSProvider says whether the name is one this server knows.
func ValidKMSProvider(name string) bool {
	for _, provider := range KMSProviders {
		if provider == name {
			return true
		}
	}
	return false
}
