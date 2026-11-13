package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/base"
)

// A provider name that is not one this server knows stops the start. Carrying on
// with the app secret would protect the stored key with something other than what
// the operator asked for, which is the one outcome worse than not starting.
func TestAnUnknownKMSProviderStopsTheStart(t *testing.T) {
	for _, name := range base.KMSProviders {
		kms := KMS{Provider: name}
		require.NoError(t, kms.validate(), name)
		assert.Equal(t, name, kms.Provider)
	}

	kms := KMS{Provider: "gcpkms"}
	err := kms.validate()
	require.ErrorIs(t, err, ErrKMSProviderUnknown)
	assert.Contains(t, err.Error(), "local", "and the message lists what it could have been")
	assert.Contains(t, err.Error(), "vault-transit")

	require.ErrorIs(t, (&KMS{Provider: "Local"}).validate(), ErrKMSProviderUnknown, "names are exact")
}

// The loader's default does not apply to a section that is present with the key
// left out, which is what a config file holding only the Vault fields looks like.
func TestAnUnsetKMSProviderIsTheAppSecret(t *testing.T) {
	kms := KMS{}
	require.NoError(t, kms.validate())
	assert.Equal(t, base.KMSProviderLocal, kms.Provider)
}
