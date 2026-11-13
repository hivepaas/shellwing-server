package kms

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/pkg/datakey"
)

// keyFrom builds a data encryption key, to tell "the same key" from "a key that
// also unwrapped".
func keyFrom(t *testing.T, raw []byte) *datakey.Key {
	t.Helper()
	key, err := datakey.New(raw)
	require.NoError(t, err)
	return key
}

func rawKey(t *testing.T) []byte {
	t.Helper()
	raw := make([]byte, datakey.KeyLen)
	_, err := rand.Read(raw)
	require.NoError(t, err)
	return raw
}

// goldenKey and goldenStored are a key and the stored value the app wrote for it
// before this package existed, with goldenSecret as the app secret. They are
// written down rather than produced here on purpose: the point is that the scheme
// has not changed, and a test that wraps and then unwraps with today's code would
// pass however much it had.
const (
	goldenSecret = "a-golden-app-secret"
	goldenKeyHex = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	goldenStored = "hpsalt:VDWiWB3Rxdf/hw== wLyhxSn7s+DuURDBMpMAzsHxqBjNZDrQAli8si3FiUcOSzJh26b62qFI0" +
		"DQr9ERd0wXQ2WAGN1vgrmeSV+L/VdX0ScBLiH4Z"
)

// An installation upgrading into this code must read the key already in its
// database, which was wrapped with the app secret and carries no marker saying
// so. If this breaks, every stored secret on every existing install is
// unreadable, so it is the one test here that matters most.
func TestTheAppSecretWrapperReadsAKeyStoredBeforeIt(t *testing.T) {
	want, err := hex.DecodeString(goldenKeyHex)
	require.NoError(t, err)
	wrapper, err := NewLocal(goldenSecret)
	require.NoError(t, err)

	got, err := wrapper.Unwrap(context.Background(), goldenStored)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, ProviderLocal, ProviderOf(goldenStored), "and it is read as the app secret's")
}

// And what it writes is still that form, so a server can be moved back to the
// app secret and an older build will read what this one stored.
func TestTheAppSecretWrapperStillWritesTheFormWithNoMarker(t *testing.T) {
	ctx := context.Background()
	raw := rawKey(t)
	wrapper, err := NewLocal(goldenSecret)
	require.NoError(t, err)

	stored, err := wrapper.Wrap(ctx, raw)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(stored, "hpsalt:"), "the form it has always had, got %q", stored)
	assert.Equal(t, ProviderLocal, ProviderOf(stored))

	// The same key, not merely one that unwraps: what it seals, a key built from
	// the bytes that went in opens.
	back, err := wrapper.Unwrap(ctx, stored)
	require.NoError(t, err)
	sealed, err := keyFrom(t, raw).Seal("a stored secret")
	require.NoError(t, err)
	opened, err := keyFrom(t, back).Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "a stored secret", opened)
}

func TestTheAppSecretWrapperRefusesTheWrongSecret(t *testing.T) {
	ctx := context.Background()
	wrapper, err := NewLocal("the-right-one")
	require.NoError(t, err)
	stored, err := wrapper.Wrap(ctx, rawKey(t))
	require.NoError(t, err)

	other, err := NewLocal("not-the-right-one")
	require.NoError(t, err)
	_, err = other.Unwrap(ctx, stored)
	require.Error(t, err)

	_, err = NewLocal("")
	require.ErrorIs(t, err, ErrNotConfigured, "an unset app secret is not a key")
}

// A stored key says which provider wrapped it, because the configuration cannot
// be trusted to still say: an operator switches provider, or restores a backup
// taken under another one.
func TestAStoredKeySaysWhoWrappedIt(t *testing.T) {
	assert.Equal(t, ProviderLocal, ProviderOf("hpsalt:c2FsdA== Y2lwaGVy"), "no marker means the app secret")
	assert.Equal(t, ProviderLocal, ProviderOf(""), "and so does nothing at all")
	assert.Equal(t, ProviderVaultTransit, ProviderOf(pack(ProviderVaultTransit, "vault:v1:abc")))
	assert.Equal(t, ProviderAWSKMS, ProviderOf(pack(ProviderAWSKMS, "AQICAHg=")))

	// A marker with nothing after it is not a marker.
	assert.Equal(t, ProviderLocal, ProviderOf("hpkms:"))
	assert.Equal(t, ProviderLocal, ProviderOf("hpkms:nocolon"))
}

func TestAWrapperRefusesAKeyAnotherProviderWrapped(t *testing.T) {
	ctx := context.Background()
	wrapper, err := NewLocal("a-secret")
	require.NoError(t, err)

	_, err = wrapper.Unwrap(ctx, pack(ProviderAWSKMS, "AQICAHg="))
	require.ErrorIs(t, err, ErrWrapped, "the app secret will not open what a service wrapped")

	_, err = payloadFor(ProviderVaultTransit, pack(ProviderAWSKMS, "AQICAHg="))
	require.ErrorIs(t, err, ErrWrapped)
	_, err = payloadFor(ProviderVaultTransit, "hpsalt:c2FsdA== Y2lwaGVy")
	require.ErrorIs(t, err, ErrWrapped)

	payload, err := payloadFor(ProviderVaultTransit, pack(ProviderVaultTransit, "vault:v1:abc"))
	require.NoError(t, err)
	assert.Equal(t, "vault:v1:abc", payload, "a payload with colons of its own survives the round trip")
}

func TestValidProvider(t *testing.T) {
	for _, name := range Providers {
		assert.True(t, ValidProvider(name), name)
	}
	assert.False(t, ValidProvider(""))
	assert.False(t, ValidProvider("gcpkms"), "not one this server knows yet")
	assert.False(t, ValidProvider("Local"), "names are exact")
}
