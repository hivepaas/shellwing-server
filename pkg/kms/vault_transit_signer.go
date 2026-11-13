package kms

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// SigningKey is a key in a service that signs, and the public half it signs under.
//
// Version is part of it, not an afterthought. A transit key can be rotated, and a
// rotated key signs under a new public half from then on (checked against Vault
// 1.20: rotating gives a v2 whose public_key differs from v1's). A certificate
// authority whose public key moved would stop being the authority every target
// server was told to trust, and every certificate already signed would be refused
// by hosts that had the new key - so a signer pins the version it was made for,
// and rotating in the service does nothing to what this server signed. Replacing
// an authority is a new key, and a new row, not a rotation underneath an old one.
type SigningKey struct {
	// Name is the key's name within the engine.
	Name string
	// Version is the key version to sign with, and whose public half PublicKey is.
	Version int
	// PublicKey is the Ed25519 public key, 32 bytes.
	PublicKey ed25519.PublicKey
}

// transitKey is what the engine says about a key.
//
// Keys is left raw because its shape depends on the kind of key: for an
// asymmetric one each version is an object with the public half in it, but for a
// symmetric one each version is a bare number, the time it was created. Decoding
// it as the object form fails on a symmetric key, which is exactly the key an
// authority must not be pointed at - so it has to be readable far enough to say
// so.
type transitKey struct {
	Data struct {
		Type          string                     `json:"type"`
		LatestVersion int                        `json:"latest_version"`
		Keys          map[string]json.RawMessage `json:"keys"`
	} `json:"data"`
}

// keyTypeEd25519 is the only type this signs with. The services that cannot do it
// - AWS, GCP and Azure KMS all do RSA and ECDSA only - need an ECDSA authority
// instead, which is a different signer and not this one.
const keyTypeEd25519 = "ed25519"

// CreateVaultTransitSigningKey makes an Ed25519 key in the engine and says what
// it is. The engine makes it unexportable (checked: "exportable": false), so this
// server can create a key it will never be able to read - which is what lets a CA
// key be made on demand without the server ever being able to leak it.
//
// Creating a name that exists does nothing in transit rather than failing, so the
// key is read back and its type checked: a name already taken by an aes256-gcm96
// key would otherwise look like success and fail later, at the first signature.
func CreateVaultTransitSigningKey(ctx context.Context, opts VaultTransitOptions) (*SigningKey, error) {
	engine, err := newVaultTransit(opts)
	if err != nil {
		return nil, err
	}
	var ignored struct{}
	if err = engine.call(ctx, "keys", map[string]string{"type": keyTypeEd25519}, &ignored); err != nil {
		return nil, err
	}
	return engine.signingKey(ctx, 0)
}

// VaultTransitSigningKey reads a key the engine already has, at one version, or
// at its latest when version is 0.
func VaultTransitSigningKey(ctx context.Context, opts VaultTransitOptions, version int) (*SigningKey, error) {
	engine, err := newVaultTransit(opts)
	if err != nil {
		return nil, err
	}
	return engine.signingKey(ctx, version)
}

// signingKey reads the key and the public half of one of its versions.
func (w *vaultTransit) signingKey(ctx context.Context, version int) (*SigningKey, error) {
	key, err := w.readKey(ctx)
	if err != nil {
		return nil, err
	}
	if key.Data.Type != keyTypeEd25519 {
		return nil, fmt.Errorf("%w: %s/keys/%s is a %s key, and an SSH authority here is %s",
			ErrKeyUnusable, w.mount, w.key, key.Data.Type, keyTypeEd25519)
	}
	if version == 0 {
		version = key.Data.LatestVersion
	}
	entry, found := key.Data.Keys[strconv.Itoa(version)]
	if !found {
		return nil, fmt.Errorf("%w: %s/keys/%s has no version %d - the version an authority "+
			"was made with cannot be deleted while it is in use",
			ErrKeyMissing, w.mount, w.key, version)
	}
	var half struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.Unmarshal(entry, &half); err != nil {
		return nil, fmt.Errorf("%w: %s version %d does not say what its public half is: %w",
			ErrServiceAnswer, w.key, version, err)
	}
	raw, err := base64.StdEncoding.DecodeString(half.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("%w: the public half of %s version %d is not base64: %w",
			ErrServiceAnswer, w.key, version, err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: the public half of %s version %d is %d bytes, not %d",
			ErrServiceAnswer, w.key, version, len(raw), ed25519.PublicKeySize)
	}
	return &SigningKey{Name: w.key, Version: version, PublicKey: ed25519.PublicKey(raw)}, nil
}

// readKey asks the engine about the key.
func (w *vaultTransit) readKey(ctx context.Context) (*transitKey, error) {
	url := fmt.Sprintf("%s/v1/%s/keys/%s", w.address, w.mount, w.key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build the %s request: %w", ProviderVaultTransit, err)
	}
	req.Header.Set("X-Vault-Token", w.token)
	if w.namespace != "" {
		req.Header.Set("X-Vault-Namespace", w.namespace)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach %s to read the key: %w", ProviderVaultTransit, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s has no key %q under %s", ErrKeyMissing,
			ProviderVaultTransit, w.key, w.mount)
	default:
		return nil, fmt.Errorf("%w: %s would not read the key %q: %s",
			ErrServiceRefused, ProviderVaultTransit, w.key, vaultError(resp))
	}

	key := &transitKey{}
	if err = json.NewDecoder(resp.Body).Decode(key); err != nil {
		return nil, fmt.Errorf("failed to read the %s answer: %w", ProviderVaultTransit, err)
	}
	return key, nil
}

// vaultSigner is a crypto.Signer whose private half is in the engine and never
// here. ssh.NewSignerFromSigner turns it into something that signs certificates.
type vaultSigner struct {
	engine *vaultTransit
	key    *SigningKey
	// ctx bounds every signature this signer makes.
	//
	// crypto.Signer.Sign takes no context and ssh.Certificate.SignCert has none
	// to give, so there is nowhere to pass one per signature. Holding the one the
	// signer was built with is what keeps a signature inside the request that
	// asked for it: when that request is abandoned, the call to the engine is too.
	ctx context.Context //nolint:containedctx // crypto.Signer has nowhere to pass one
}

// NewVaultTransitSigner signs with a key the engine holds, at the version given.
//
// The version is required rather than defaulted to the latest: signing under
// whatever the engine most recently rotated to would change the authority's
// public key without anything here saying so. VaultTransitSigningKey is how a
// caller finds the version in the first place.
func NewVaultTransitSigner(ctx context.Context, opts VaultTransitOptions,
	key *SigningKey) (crypto.Signer, error) {
	engine, err := newVaultTransit(opts)
	if err != nil {
		return nil, err
	}
	switch {
	case key == nil || key.Version <= 0:
		return nil, fmt.Errorf("%w: signing needs the key version it was made for", ErrNotConfigured)
	case len(key.PublicKey) != ed25519.PublicKeySize:
		return nil, fmt.Errorf("%w: signing needs the key's public half", ErrNotConfigured)
	}
	return &vaultSigner{engine: engine, key: key, ctx: ctx}, nil
}

// Public is the half the engine signs under, which a crypto.Signer has to be able
// to answer without asking anybody.
func (s *vaultSigner) Public() crypto.PublicKey {
	return s.key.PublicKey
}

// Sign signs the message with the key in the engine.
//
// message is the message itself, not a digest: x/crypto/ssh hands an Ed25519
// signer the bytes to sign with crypto.Hash(0), because Ed25519 hashes them
// itself. transit agrees, and refuses a hash_algorithm for an Ed25519 key with
// "only Pure Ed25519 signatures supported", so nothing is set here and a caller
// that wanted a prehashed signature is turned away rather than quietly given a
// signature over its digest.
func (s *vaultSigner) Sign(_ io.Reader, message []byte, opts crypto.SignerOpts) ([]byte, error) {
	if opts != nil && opts.HashFunc() != 0 {
		return nil, fmt.Errorf("%w: %s signs pure Ed25519, and was asked for %s",
			ErrKeyUnusable, ProviderVaultTransit, opts.HashFunc())
	}
	var answer struct {
		Data struct {
			Signature  string `json:"signature"`
			KeyVersion int    `json:"key_version"`
		} `json:"data"`
	}
	request := map[string]any{
		"input":       base64.StdEncoding.EncodeToString(message),
		"key_version": s.key.Version,
	}
	if err := s.engine.call(s.ctx, "sign", request, &answer); err != nil {
		return nil, err
	}
	if answer.Data.KeyVersion != s.key.Version {
		return nil, fmt.Errorf("%w: asked %s to sign with version %d and it signed with %d",
			ErrServiceAnswer, ProviderVaultTransit, s.key.Version, answer.Data.KeyVersion)
	}
	signature, err := decodeVaultSignature(answer.Data.Signature)
	if err != nil {
		return nil, err
	}
	// Checked here rather than left to whoever verifies: a signature of the wrong
	// length would otherwise become a certificate that no server accepts, found by
	// somebody who cannot log in.
	if len(signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("%w: %s returned a %d byte signature, not %d",
			ErrServiceAnswer, ProviderVaultTransit, len(signature), ed25519.SignatureSize)
	}
	return signature, nil
}

// decodeVaultSignature reads what transit answers with, which is the marker, the
// key version that made it, and the signature: vault:v1:<base64>.
func decodeVaultSignature(signature string) ([]byte, error) {
	marker, rest, found := strings.Cut(signature, ":")
	if !found || marker != "vault" {
		return nil, fmt.Errorf("%w: %s returned %q, which is not a signature",
			ErrServiceAnswer, ProviderVaultTransit, signature)
	}
	_, encoded, found := strings.Cut(rest, ":")
	if !found {
		return nil, fmt.Errorf("%w: %s returned a signature with no version in it",
			ErrServiceAnswer, ProviderVaultTransit)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %s returned a signature that is not base64: %w",
			ErrServiceAnswer, ProviderVaultTransit, err)
	}
	return raw, nil
}
