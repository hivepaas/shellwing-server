package identity

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// secret is 32 bytes counting up from from: the vector's secrets are 1, 2, 3… for X25519 and 33, 34, 35… for Ed25519.
func secret(from byte) []byte {
	b := make([]byte, KeyLen)
	for i := range b {
		b[i] = from + byte(i)
	}
	return b
}

func b64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return b
}

// The vector the app checks too (shellwing's crates/vault/src/identity.rs), worked out again here from the secrets.
func TestTheDocumentedVector(t *testing.T) {
	x, err := ecdh.X25519().NewPrivateKey(secret(1))
	require.NoError(t, err)
	edSecret := ed25519.NewKeyFromSeed(secret(33))
	xPub, edPub := x.PublicKey().Bytes(), []byte(edSecret.Public().(ed25519.PublicKey))
	signature := ed25519.Sign(edSecret, signed(xPub, edPub))

	assert.Equal(t, b64(t, "B6N8vBQgk8i3VdwbEOhstCY3StFqqFPtC9/AsrhtHHw="), xPub)
	assert.Equal(t, b64(t, "5/FioQvsVZr+oZXk3OhLaVaNXSywlj60RsBoXisX8vA="), edPub)
	assert.Equal(t, b64(t, "I+mHldAuIZ4mZvZQbMr+4cdL6LdTZ7qA1Xz0egaIf0MvLbChGT55sA1gZQTCnuJV+hW3ZzeKafF1lHrxk2uNCA=="),
		signature)
	assert.True(t, Verify(xPub, edPub, signature))
	assert.Equal(t, "42092 42879 96924 77795 39558 81840", Fingerprint(xPub, edPub))
}

func TestAnotherKeyIsNotSigned(t *testing.T) {
	xPub := b64(t, "B6N8vBQgk8i3VdwbEOhstCY3StFqqFPtC9/AsrhtHHw=")
	edPub := b64(t, "5/FioQvsVZr+oZXk3OhLaVaNXSywlj60RsBoXisX8vA=")
	signature := b64(t, "I+mHldAuIZ4mZvZQbMr+4cdL6LdTZ7qA1Xz0egaIf0MvLbChGT55sA1gZQTCnuJV+hW3ZzeKafF1lHrxk2uNCA==")
	forged := append([]byte(nil), xPub...)
	forged[0] ^= 1
	assert.False(t, Verify(forged, edPub, signature))
	assert.False(t, Verify(xPub[:31], edPub, signature), "a short key")
	assert.False(t, Verify(xPub, edPub, signature[:63]), "a short signature")
}
