package sshca

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/hperrors"
)

func userKey(t *testing.T) (ssh.PublicKey, ssh.Signer) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPublic, err := ssh.NewPublicKey(public)
	require.NoError(t, err)
	signer, err := ssh.NewSignerFromKey(private)
	require.NoError(t, err)
	return sshPublic, signer
}

func authorized(t *testing.T, key any) string {
	t.Helper()
	sshPublic, err := ssh.NewPublicKey(key)
	require.NoError(t, err)
	return string(ssh.MarshalAuthorizedKey(sshPublic))
}

func parsedCert(t *testing.T, text string) *ssh.Certificate {
	t.Helper()
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(text)) //nolint:dogsled // only the key
	require.NoError(t, err)
	cert, ok := key.(*ssh.Certificate)
	require.True(t, ok)
	return cert
}

func TestSignMakesACertificateSSHTrusts(t *testing.T) {
	ca, err := GenerateCA()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(ca.PublicKey, "ssh-ed25519 "))
	caPublic, _, _, _, err := ssh.ParseAuthorizedKey([]byte(ca.PublicKey)) //nolint:dogsled // only the key
	require.NoError(t, err)

	key, _ := userKey(t)
	now := time.Now().UTC()
	signed, err := Sign(ca.PrivateKey, &UserCert{
		Key: key, KeyID: "user=alice", Principals: []string{"deploy", "alice"}, TTL: time.Hour, Now: now,
	})
	require.NoError(t, err)
	cert := parsedCert(t, signed.Certificate)

	assert.Equal(t, uint32(ssh.UserCert), cert.CertType)
	assert.Equal(t, "user=alice", cert.KeyId)
	assert.Equal(t, []string{"deploy", "alice"}, cert.ValidPrincipals)
	assert.Equal(t, ssh.FingerprintSHA256(key), signed.Fingerprint)
	assert.Equal(t, signed.Serial, cert.Serial)
	assert.Empty(t, cert.CriticalOptions, "a critical option an sshd does not know refuses the login")

	checker := &ssh.CertChecker{
		IsUserAuthority: func(auth ssh.PublicKey) bool {
			return string(auth.Marshal()) == string(caPublic.Marshal())
		},
		Clock: func() time.Time { return now },
	}
	require.NoError(t, checker.CheckCert("deploy", cert))
	require.Error(t, checker.CheckCert("root", cert), "only the accounts it names")

	checker.Clock = func() time.Time { return now.Add(time.Hour + time.Minute) }
	require.Error(t, checker.CheckCert("deploy", cert), "it ends")
	checker.Clock = func() time.Time { return now.Add(-time.Minute) }
	require.NoError(t, checker.CheckCert("deploy", cert), "a server whose clock is a little behind takes it")
}

func TestSignRefusesACertificateWithNoPrincipals(t *testing.T) {
	ca, err := GenerateCA()
	require.NoError(t, err)
	key, _ := userKey(t)
	_, err = Sign(ca.PrivateKey, &UserCert{Key: key, TTL: time.Hour, Now: time.Now()})
	require.ErrorIs(t, err, hperrors.ErrSSHNoPrincipals, "sshd reads no principals as any account")
}

func TestParseUserKey(t *testing.T) {
	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, err = ParseUserKey(authorized(t, public) + "\n")
	require.NoError(t, err)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	_, err = ParseUserKey(authorized(t, &ecKey.PublicKey))
	require.NoError(t, err)

	bigRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	_, err = ParseUserKey(authorized(t, &bigRSA.PublicKey))
	require.NoError(t, err)

	smallRSA, err := rsa.GenerateKey(rand.Reader, 1024) //nolint:gosec // G403: the small key under test
	require.NoError(t, err)
	_, err = ParseUserKey(authorized(t, &smallRSA.PublicKey))
	require.ErrorIs(t, err, hperrors.ErrSSHPublicKeyInvalid)

	_, err = ParseUserKey("not a key")
	require.ErrorIs(t, err, hperrors.ErrSSHPublicKeyInvalid)

	ca, err := GenerateCA()
	require.NoError(t, err)
	key, _ := userKey(t)
	signed, err := Sign(ca.PrivateKey, &UserCert{Key: key, Principals: []string{"a"}, TTL: time.Hour, Now: time.Now()})
	require.NoError(t, err)
	_, err = ParseUserKey(signed.Certificate)
	require.ErrorIs(t, err, hperrors.ErrSSHPublicKeyInvalid, "a certificate of a certificate")
}

func TestValidPrincipal(t *testing.T) {
	for _, name := range []string{"deploy", "root", "alice.b", "svc-web", "a@corp.example", "_ops", "1"} {
		assert.True(t, ValidPrincipal(name), name)
	}
	for _, name := range []string{"", "a,b", "a b", "#a", "-a", ".a", "a\n", "a\nb", strings.Repeat("a", 65), "*"} {
		assert.False(t, ValidPrincipal(name), name)
	}
}
