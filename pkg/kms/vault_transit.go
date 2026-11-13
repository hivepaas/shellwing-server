package kms

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// defaultVaultTimeout bounds a call to the transit engine. Wrapping happens at
// startup, so a server whose Vault is unreachable should say so rather than hang
// where nobody is watching.
const defaultVaultTimeout = 10 * time.Second

// defaultVaultMount is where the transit engine sits unless it was mounted
// elsewhere.
const defaultVaultMount = "transit"

// VaultTransitOptions is what reaching a transit engine needs.
type VaultTransitOptions struct {
	// Address is the base URL, such as https://vault.internal:8200.
	Address string
	// Token authenticates this server. It is a credential of the server's, not of
	// any account's, and revoking it is what retires every backup of the database
	// at once.
	Token string
	// Mount is the engine's path, "transit" unless it was mounted elsewhere.
	Mount string
	// Key is the name of the key within the engine. The engine never hands it
	// over; it only answers with what the key opened.
	Key string
	// Namespace is Vault Enterprise's namespace, if it is in one. OpenBao and
	// Vault's community edition have none.
	Namespace string
	// CACertFile is a PEM file to trust for the connection, for the private
	// authority a self-hosted Vault is usually behind. Empty, the system's store
	// is used.
	CACertFile string
	// Timeout bounds one call. Zero is defaultVaultTimeout.
	Timeout time.Duration
}

// vaultTransit wraps the data encryption key with Vault's or OpenBao's transit
// engine.
//
// Transit is the right engine for this rather than a KV secret: a KV secret
// would mean fetching the key encryption key and holding it here, which is the
// thing a KMS is for avoiding. Transit keeps the key and answers with the result.
type vaultTransit struct {
	address   string
	token     string
	mount     string
	key       string
	namespace string
	client    *http.Client
}

// NewVaultTransit reaches the transit engine of a HashiCorp Vault or an OpenBao.
// It does not call it: whether the token works is found on the first wrap or
// unwrap, which is at startup.
func NewVaultTransit(opts VaultTransitOptions) (Wrapper, error) {
	switch {
	case opts.Address == "":
		return nil, fmt.Errorf("%w: %s needs an address", ErrNotConfigured, ProviderVaultTransit)
	case opts.Token == "":
		return nil, fmt.Errorf("%w: %s needs a token", ErrNotConfigured, ProviderVaultTransit)
	case opts.Key == "":
		return nil, fmt.Errorf("%w: %s needs a key name", ErrNotConfigured, ProviderVaultTransit)
	}

	transport, err := vaultTransport(opts.CACertFile)
	if err != nil {
		return nil, err
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultVaultTimeout
	}
	mount := opts.Mount
	if mount == "" {
		mount = defaultVaultMount
	}
	return &vaultTransit{
		address:   strings.TrimRight(opts.Address, "/"),
		token:     opts.Token,
		mount:     strings.Trim(mount, "/"),
		key:       opts.Key,
		namespace: opts.Namespace,
		client:    &http.Client{Timeout: timeout, Transport: transport},
	}, nil
}

// vaultTransport trusts the given authority instead of the system's, for the
// private CA a self-hosted Vault is usually behind.
func vaultTransport(caCertFile string) (http.RoundTripper, error) {
	if caCertFile == "" {
		return nil, nil //nolint:nilnil // the default transport, which is what nil means here
	}
	pem, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read the CA certificate %s: %w",
			ErrNotConfigured, caCertFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("%w: %s holds no certificate", ErrNotConfigured, caCertFile)
	}
	return &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}, nil
}

func (w *vaultTransit) Name() string { return ProviderVaultTransit }

func (w *vaultTransit) Wrap(ctx context.Context, key []byte) (string, error) {
	var answer struct {
		Data struct {
			Ciphertext string `json:"ciphertext"`
		} `json:"data"`
	}
	request := map[string]string{"plaintext": base64.StdEncoding.EncodeToString(key)}
	if err := w.call(ctx, "encrypt", request, &answer); err != nil {
		return "", err
	}
	if answer.Data.Ciphertext == "" {
		return "", fmt.Errorf("%w: %s returned none", ErrServiceAnswer, ProviderVaultTransit)
	}
	return pack(ProviderVaultTransit, answer.Data.Ciphertext), nil
}

func (w *vaultTransit) Unwrap(ctx context.Context, stored string) ([]byte, error) {
	ciphertext, err := payloadFor(ProviderVaultTransit, stored)
	if err != nil {
		return nil, err
	}
	var answer struct {
		Data struct {
			Plaintext string `json:"plaintext"`
		} `json:"data"`
	}
	if err = w.call(ctx, "decrypt", map[string]string{"ciphertext": ciphertext}, &answer); err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(answer.Data.Plaintext)
	if err != nil {
		return nil, fmt.Errorf("%w: %s returned a plaintext that is not base64: %w",
			ErrWrapped, ProviderVaultTransit, err)
	}
	return key, nil
}

// call posts to one of the engine's endpoints and reads its answer.
func (w *vaultTransit) call(ctx context.Context, action string, request, answer any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to encode the %s request: %w", ProviderVaultTransit, err)
	}
	url := fmt.Sprintf("%s/v1/%s/%s/%s", w.address, w.mount, action, w.key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build the %s request: %w", ProviderVaultTransit, err)
	}
	req.Header.Set("X-Vault-Token", w.token)
	req.Header.Set("Content-Type", "application/json")
	if w.namespace != "" {
		req.Header.Set("X-Vault-Namespace", w.namespace)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		// The URL is in the error already; the token is not, and must not be.
		return fmt.Errorf("cannot reach %s to %s the data encryption key: %w",
			ProviderVaultTransit, action, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s would not %s the data encryption key: %s",
			ErrServiceRefused, ProviderVaultTransit, action, vaultError(resp))
	}
	if err = json.NewDecoder(resp.Body).Decode(answer); err != nil {
		return fmt.Errorf("failed to read the %s answer: %w", ProviderVaultTransit, err)
	}
	return nil
}

// vaultError reads what the engine said went wrong, for a message that names the
// cause rather than only the status.
func vaultError(resp *http.Response) string {
	var answer struct {
		Errors []string `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&answer); err == nil && len(answer.Errors) > 0 {
		return fmt.Sprintf("%s (%s)", resp.Status, strings.Join(answer.Errors, "; "))
	}
	return resp.Status
}
