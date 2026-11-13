package kms

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// A key made for this run, so a test never leans on one an earlier run left.
func liveSigningKeyName(t *testing.T) string {
	t.Helper()
	return "shellwing-test-ca-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// The one that matters: a certificate signed by a key this server cannot read,
// accepted by the library that speaks for every sshd.
func TestALiveTransitKeySignsACertificateSSHTrusts(t *testing.T) {
	ctx := context.Background()
	opts := liveVault(t, liveSigningKeyName(t))

	key, err := CreateVaultTransitSigningKey(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, 1, key.Version, "a key just made is at its first version")
	assert.Len(t, key.PublicKey, ed25519.PublicKeySize)

	signer, err := NewVaultTransitSigner(ctx, opts, key)
	require.NoError(t, err)
	sshSigner, err := ssh.NewSignerFromSigner(signer)
	require.NoError(t, err)
	assert.Equal(t, ssh.KeyAlgoED25519, sshSigner.PublicKey().Type(),
		"an ordinary ssh-ed25519 authority, so TrustedUserCAKeys takes it unchanged")

	userPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	userKey, err := ssh.NewPublicKey(userPublic)
	require.NoError(t, err)
	now := time.Now()
	cert := &ssh.Certificate{
		Key:             userKey,
		Serial:          1,
		CertType:        ssh.UserCert,
		KeyId:           "user=alice",
		ValidPrincipals: []string{"deploy"},
		ValidAfter:      uint64(now.Add(-time.Minute).Unix()),
		ValidBefore:     uint64(now.Add(time.Hour).Unix()),
	}
	require.NoError(t, cert.SignCert(rand.Reader, sshSigner))

	checker := &ssh.CertChecker{IsUserAuthority: func(auth ssh.PublicKey) bool {
		return string(auth.Marshal()) == string(sshSigner.PublicKey().Marshal())
	}}
	require.NoError(t, checker.CheckCert("deploy", cert),
		"signed in the engine, verified here, with the private half never in this process")
}

// Rotating the key in the engine must not move the authority under the hosts that
// were told to trust it. The signer is pinned to the version it was made for.
func TestALiveTransitKeyKeepsSigningUnderTheVersionItWasMadeWith(t *testing.T) {
	ctx := context.Background()
	opts := liveVault(t, liveSigningKeyName(t))

	first, err := CreateVaultTransitSigningKey(ctx, opts)
	require.NoError(t, err)
	signer, err := NewVaultTransitSigner(ctx, opts, first)
	require.NoError(t, err)

	// Rotate, as an operator or an auto_rotate_period would.
	require.NoError(t, rotateLive(ctx, opts))

	latest, err := VaultTransitSigningKey(ctx, opts, 0)
	require.NoError(t, err)
	require.Greater(t, latest.Version, first.Version, "the engine has a newer version now")
	assert.NotEqual(t, first.PublicKey, latest.PublicKey, "and it signs under a different public half")

	// The signer made before the rotation is unmoved.
	assert.Equal(t, first.PublicKey, signer.Public(),
		"the authority's public key did not change under it")
	message := []byte("after the rotation")
	signature, err := signer.Sign(rand.Reader, message, crypto.Hash(0))
	require.NoError(t, err)
	assert.True(t, ed25519.Verify(first.PublicKey, message, signature),
		"and what it signs still verifies against the half the hosts trust")
	assert.False(t, ed25519.Verify(latest.PublicKey, message, signature))
}

// rotateLive rotates the key. It posts directly because the rotate path is
// keys/<name>/rotate, which is not the shape the engine's other calls take, and
// rotating is something only a test here does.
func rotateLive(ctx context.Context, opts VaultTransitOptions) error {
	url := opts.Address + "/v1/transit/keys/" + opts.Key + "/rotate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", opts.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("rotating %s: %s", opts.Key, resp.Status)
	}
	return nil
}

// A name already taken by a key of another kind is the mistake worth catching:
// transit answers a create for an existing name with the key it already has, and
// does not change its type, so nothing but a read tells.
func TestALiveTransitKeyOfTheWrongKindIsRefused(t *testing.T) {
	ctx := context.Background()
	opts := liveVault(t, "shellwing-dek") // the aes256-gcm96 key the KEK uses

	_, err := CreateVaultTransitSigningKey(ctx, opts)
	require.ErrorIs(t, err, ErrKeyUnusable)
	assert.Contains(t, err.Error(), "aes256-gcm96")
	assert.Contains(t, err.Error(), "ed25519", "and says what an authority has to be")
}

func TestALiveTransitSignerNeedsAVersionThatIsThere(t *testing.T) {
	ctx := context.Background()
	opts := liveVault(t, liveSigningKeyName(t))
	_, err := CreateVaultTransitSigningKey(ctx, opts)
	require.NoError(t, err)

	_, err = VaultTransitSigningKey(ctx, opts, 99)
	require.ErrorIs(t, err, ErrKeyMissing)
	assert.Contains(t, err.Error(), "99")
}

func TestTheTransitSignerNeedsTheKeyItSignsUnder(t *testing.T) {
	ctx := context.Background()
	opts := VaultTransitOptions{Address: "https://vault:8200", Token: "t", Key: "k"}

	_, err := NewVaultTransitSigner(ctx, opts, nil)
	require.ErrorIs(t, err, ErrNotConfigured)

	_, err = NewVaultTransitSigner(ctx, opts, &SigningKey{Version: 0,
		PublicKey: make([]byte, ed25519.PublicKeySize)})
	require.ErrorIs(t, err, ErrNotConfigured, "signing under the latest is not a thing it does")

	_, err = NewVaultTransitSigner(ctx, opts, &SigningKey{Version: 1, PublicKey: []byte("short")})
	require.ErrorIs(t, err, ErrNotConfigured)
}

// What the engine answers with is checked rather than trusted: a signature of the
// wrong length, or made under another version, would become a certificate no
// server accepts, found by somebody who cannot log in.
func TestTheTransitSignerChecksWhatCameBack(t *testing.T) {
	ctx := context.Background()
	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key := &SigningKey{Name: "k", Version: 2, PublicKey: public}

	for _, c := range []struct {
		name   string
		answer string
		says   string
	}{
		{"another version", `{"data":{"signature":"vault:v3:` + sixtyFour + `","key_version":3}}`,
			"signed with 3"},
		{"too short", `{"data":{"signature":"vault:v2:Zm9v","key_version":2}}`, "3 byte signature"},
		{"not a signature", `{"data":{"signature":"nonsense","key_version":2}}`, "not a signature"},
		{"no version in it", `{"data":{"signature":"vault:","key_version":2}}`, "no version in it"},
		{"not base64", `{"data":{"signature":"vault:v2:not base64!!","key_version":2}}`, "not base64"},
	} {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(c.answer))
			}))
			defer server.Close()

			signer, err := NewVaultTransitSigner(ctx,
				VaultTransitOptions{Address: server.URL, Token: "t", Key: "k"}, key)
			require.NoError(t, err)
			_, err = signer.Sign(rand.Reader, []byte("message"), crypto.Hash(0))
			require.ErrorIs(t, err, ErrServiceAnswer)
			assert.Contains(t, err.Error(), c.says)
		})
	}
}

// transit refuses a hash_algorithm for an Ed25519 key, so a caller that wanted a
// signature over a digest is turned away here rather than quietly given one over
// the digest bytes themselves.
func TestTheTransitSignerRefusesAPrehashedSignature(t *testing.T) {
	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	signer, err := NewVaultTransitSigner(context.Background(),
		VaultTransitOptions{Address: "https://vault:8200", Token: "t", Key: "k"},
		&SigningKey{Name: "k", Version: 1, PublicKey: public})
	require.NoError(t, err)

	_, err = signer.Sign(rand.Reader, []byte("digest"), crypto.SHA512)
	require.ErrorIs(t, err, ErrKeyUnusable)
	assert.Contains(t, err.Error(), "SHA-512")

	assert.Equal(t, public, signer.Public(), "and Public needs nobody")
}

// sixtyFour is base64 of 64 zero bytes: a signature of the right length.
const sixtyFour = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
