package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vaultCall is what the engine was asked to do, so a test can check the request
// as well as what came back from it.
type vaultCall struct {
	method    string
	path      string
	token     string
	namespace string
	body      map[string]string
}

// fakeVault answers as the transit engine does, and records what it was asked.
func fakeVault(t *testing.T, answer func(call *vaultCall) (int, string)) (*httptest.Server, *[]vaultCall) {
	t.Helper()
	calls := &[]vaultCall{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		call := vaultCall{
			method:    r.Method,
			path:      r.URL.Path,
			token:     r.Header.Get("X-Vault-Token"),
			namespace: r.Header.Get("X-Vault-Namespace"),
		}
		require.NoError(t, json.Unmarshal(raw, &call.body))
		*calls = append(*calls, call)

		status, payload := answer(&call)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(server.Close)
	return server, calls
}

func TestTheTransitEngineWrapsAndUnwrapsTheKey(t *testing.T) {
	ctx := context.Background()
	raw := rawKey(t)
	var held string // what the engine was given, as it would hold it

	server, calls := fakeVault(t, func(call *vaultCall) (int, string) {
		if call.body["plaintext"] != "" {
			held = call.body["plaintext"]
			return http.StatusOK, `{"data":{"ciphertext":"vault:v1:Zm9v"}}`
		}
		assert.Equal(t, "vault:v1:Zm9v", call.body["ciphertext"], "it gives back what it was given")
		return http.StatusOK, `{"data":{"plaintext":"` + held + `"}}`
	})

	wrapper, err := NewVaultTransit(VaultTransitOptions{
		Address: server.URL, Token: "a-server-token", Key: "shellwing-dek", Namespace: "team-a",
	})
	require.NoError(t, err)
	assert.Equal(t, ProviderVaultTransit, wrapper.Name())

	stored, err := wrapper.Wrap(ctx, raw)
	require.NoError(t, err)
	assert.Equal(t, ProviderVaultTransit, ProviderOf(stored), "the stored key says who wrapped it")
	assert.Equal(t, pack(ProviderVaultTransit, "vault:v1:Zm9v"), stored)

	back, err := wrapper.Unwrap(ctx, stored)
	require.NoError(t, err)
	assert.Equal(t, raw, back)

	require.Len(t, *calls, 2)
	wrap, unwrap := (*calls)[0], (*calls)[1]
	assert.Equal(t, http.MethodPost, wrap.method)
	assert.Equal(t, "/v1/transit/encrypt/shellwing-dek", wrap.path, "the default mount, and the key's name")
	assert.Equal(t, "/v1/transit/decrypt/shellwing-dek", unwrap.path)
	assert.Equal(t, "a-server-token", wrap.token, "the token goes in the header, not the URL")
	assert.Equal(t, "team-a", wrap.namespace)
	assert.Equal(t, base64.StdEncoding.EncodeToString(raw), wrap.body["plaintext"],
		"the key is sent base64, which is what transit takes")
}

func TestTheTransitEngineUsesTheMountItWasGiven(t *testing.T) {
	server, calls := fakeVault(t, func(*vaultCall) (int, string) {
		return http.StatusOK, `{"data":{"ciphertext":"vault:v1:Zm9v"}}`
	})
	wrapper, err := NewVaultTransit(VaultTransitOptions{
		Address: server.URL + "/", Token: "t", Mount: "/kms-transit/", Key: "k",
	})
	require.NoError(t, err)
	_, err = wrapper.Wrap(context.Background(), rawKey(t))
	require.NoError(t, err)

	require.Len(t, *calls, 1)
	assert.Equal(t, "/v1/kms-transit/encrypt/k", (*calls)[0].path,
		"a trailing slash on either does not double up")
	assert.Empty(t, (*calls)[0].namespace, "no namespace header when there is no namespace")
}

// A refusal has to say why. A sealed Vault and a revoked token both answer with a
// status and a message, and the message is the half that tells them apart.
func TestTheTransitEngineSaysWhyItRefused(t *testing.T) {
	server, _ := fakeVault(t, func(*vaultCall) (int, string) {
		return http.StatusForbidden, `{"errors":["permission denied"]}`
	})
	wrapper, err := NewVaultTransit(VaultTransitOptions{Address: server.URL, Token: "stale", Key: "k"})
	require.NoError(t, err)

	_, err = wrapper.Wrap(context.Background(), rawKey(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	assert.Contains(t, err.Error(), "403")
	assert.NotContains(t, err.Error(), "stale", "and never the token")
}

func TestTheTransitEngineRefusesAnAnswerItCannotUse(t *testing.T) {
	ctx := context.Background()
	server, _ := fakeVault(t, func(call *vaultCall) (int, string) {
		if call.body["plaintext"] != "" {
			return http.StatusOK, `{"data":{}}`
		}
		return http.StatusOK, `{"data":{"plaintext":"not base64 at all!!"}}`
	})
	wrapper, err := NewVaultTransit(VaultTransitOptions{Address: server.URL, Token: "t", Key: "k"})
	require.NoError(t, err)

	_, err = wrapper.Wrap(ctx, rawKey(t))
	require.Error(t, err, "no ciphertext is not a wrap")

	_, err = wrapper.Unwrap(ctx, pack(ProviderVaultTransit, "vault:v1:Zm9v"))
	require.ErrorIs(t, err, ErrWrapped)
}

func TestTheTransitEngineNeedsEnoughToReachIt(t *testing.T) {
	for _, missing := range []struct {
		name string
		opts VaultTransitOptions
	}{
		{"no address", VaultTransitOptions{Token: "t", Key: "k"}},
		{"no token", VaultTransitOptions{Address: "https://vault:8200", Key: "k"}},
		{"no key", VaultTransitOptions{Address: "https://vault:8200", Token: "t"}},
	} {
		_, err := NewVaultTransit(missing.opts)
		require.ErrorIs(t, err, ErrNotConfigured, missing.name)
	}

	_, err := NewVaultTransit(VaultTransitOptions{
		Address: "https://vault:8200", Token: "t", Key: "k", CACertFile: "/nope/no-such.pem",
	})
	require.ErrorIs(t, err, ErrNotConfigured, "a CA file that is not there is a configuration error")
}
