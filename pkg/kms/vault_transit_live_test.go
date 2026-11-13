package kms

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveVault returns the options for a real Vault or OpenBao, or skips.
//
// The other tests here answer to an httptest double, which proves this code does
// what its author thinks transit does. It cannot prove transit does it. Point
// SHW_TEST_VAULT_ADDR and SHW_TEST_VAULT_TOKEN at a dev-mode server with the
// transit engine mounted to check that too:
//
//	docker run -d --name vault -p 127.0.0.1:18200:8200 \
//	  -e VAULT_DEV_ROOT_TOKEN_ID=test-root-token --cap-add=IPC_LOCK \
//	  hashicorp/vault:1.20 server -dev -dev-listen-address=0.0.0.0:8200
//	curl -X POST -H 'X-Vault-Token: test-root-token' -d '{"type":"transit"}' \
//	  http://127.0.0.1:18200/v1/sys/mounts/transit
//	curl -X POST -H 'X-Vault-Token: test-root-token' \
//	  http://127.0.0.1:18200/v1/transit/keys/shellwing-dek
//
// CI has no Vault, so these skip there.
func liveVault(t *testing.T, key string) VaultTransitOptions {
	t.Helper()
	address, token := os.Getenv("SHW_TEST_VAULT_ADDR"), os.Getenv("SHW_TEST_VAULT_TOKEN")
	if address == "" || token == "" {
		t.Skip("SHW_TEST_VAULT_ADDR and SHW_TEST_VAULT_TOKEN not set, skipping the live transit test")
	}
	return VaultTransitOptions{Address: address, Token: token, Key: key}
}

func TestALiveTransitEngineWrapsAndUnwrapsTheKey(t *testing.T) {
	ctx := context.Background()
	wrapper, err := NewVaultTransit(liveVault(t, "shellwing-dek"))
	require.NoError(t, err)

	raw := rawKey(t)
	stored, err := wrapper.Wrap(ctx, raw)
	require.NoError(t, err)
	assert.Equal(t, ProviderVaultTransit, ProviderOf(stored))

	back, err := wrapper.Unwrap(ctx, stored)
	require.NoError(t, err)
	assert.Equal(t, raw, back, "what a real transit engine gives back is the key that went in")

	// Twice over: transit's ciphertext carries a nonce, so the same key wraps to
	// two different values, and both have to open.
	again, err := wrapper.Wrap(ctx, raw)
	require.NoError(t, err)
	assert.NotEqual(t, stored, again, "transit does not write the same ciphertext twice")
	back, err = wrapper.Unwrap(ctx, again)
	require.NoError(t, err)
	assert.Equal(t, raw, back)
}

func TestALiveTransitEngineRefusesAWrongToken(t *testing.T) {
	ctx := context.Background()
	opts := liveVault(t, "shellwing-dek")
	opts.Token = "not-the-token"
	wrapper, err := NewVaultTransit(opts)
	require.NoError(t, err)

	_, err = wrapper.Wrap(ctx, rawKey(t))
	require.ErrorIs(t, err, ErrServiceRefused)
	assert.Contains(t, err.Error(), "permission denied", "and says what transit said")
	assert.NotContains(t, err.Error(), "not-the-token")
}

// The one that found this: transit's encrypt endpoint creates a key it does not
// know. Wrap looks the key up first, so a mistyped name is refused rather than
// quietly given a new key. The name is unique to the run, because a key an older
// build auto-created would still be sitting there.
func TestALiveTransitEngineRefusesAKeyThatIsNotThere(t *testing.T) {
	ctx := context.Background()
	absent := "shellwing-no-such-key-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	wrapper, err := NewVaultTransit(liveVault(t, absent))
	require.NoError(t, err)

	_, err = wrapper.Wrap(ctx, rawKey(t))
	require.ErrorIs(t, err, ErrKeyMissing)
	assert.Contains(t, err.Error(), absent)

	// And it was not made in the asking.
	opts := liveVault(t, absent)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		opts.Address+"/v1/transit/keys/"+absent, nil)
	require.NoError(t, err)
	req.Header.Set("X-Vault-Token", opts.Token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "transit still has no such key")
}
