// Package kms wraps the data encryption key with a key held where this server
// cannot read it.
//
// Stored secrets are sealed with a data encryption key (see pkg/datakey), and
// that key lives in the database wrapped by a key encryption key. What holds the
// KEK is this package's subject: the app secret from the configuration, which is
// the default and needs nothing deployed beside the server, or a key management
// service that never hands its key over and only answers with the unwrapped DEK.
//
// What a KMS buys here is that the database, and a backup of it, stop being
// enough. Whoever takes both still has to be able to call the KMS as this server,
// which is a separate credential with its own audit trail and its own
// revocation, and revoking it retires every copy of the backup at once.
//
// What it does not buy: anything against somebody already running as this
// process. The DEK is in its memory, and so is an organization's SSH CA key for
// as long as it signs with it. Keeping the CA key out of this process altogether
// is a different seam - a crypto.Signer backed by the KMS, in pkg/sshca - and
// this package is the half that comes first because it covers every stored
// secret rather than one kind.
package kms

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hivepaas/shellwing-server/base"
)

// The provider names, re-exported from base so that this package reads as its
// own: base holds them because config validates the configured name and cannot
// import this package.
const (
	ProviderLocal        = base.KMSProviderLocal
	ProviderVaultTransit = base.KMSProviderVaultTransit
	ProviderAWSKMS       = base.KMSProviderAWSKMS
)

// Providers are the names a configuration may give, in the order a message
// listing them should read.
var Providers = base.KMSProviders

// ValidProvider says whether the name is one this server knows.
func ValidProvider(name string) bool {
	return base.ValidKMSProvider(name)
}

var (
	// ErrProviderUnknown is a name that is not one of Providers.
	ErrProviderUnknown = errors.New("unknown KMS provider")
	// ErrNotConfigured is a provider named without what it needs to reach its
	// service.
	ErrNotConfigured = errors.New("KMS provider is not configured")
	// ErrWrapped is a stored key that does not read as one this package wrote.
	ErrWrapped = errors.New("stored key is not readable")
	// ErrServiceRefused is the service declining the call: a sealed Vault, a
	// revoked token, a key policy that does not admit this caller.
	ErrServiceRefused = errors.New("KMS refused the request")
	// ErrServiceAnswer is an answer that cannot be used - no ciphertext where one
	// was asked for - which is a service behaving unlike itself rather than
	// refusing, and is worth telling apart from a refusal.
	ErrServiceAnswer = errors.New("KMS returned an answer that cannot be used")
)

// Wrapper wraps and unwraps the data encryption key. A call reaches a service
// over the network for every provider but local, so each takes a context and may
// fail for reasons that have nothing to do with the key.
type Wrapper interface {
	// Name is the provider's name, as Providers has it.
	Name() string
	// Wrap seals the key for storing. The result records which provider made it,
	// so a later start can tell.
	Wrap(ctx context.Context, key []byte) (string, error)
	// Unwrap recovers a key that Wrap sealed.
	Unwrap(ctx context.Context, stored string) ([]byte, error)
}

// storedPrefix marks a key wrapped by a service rather than by the app secret.
// The provider's name follows it, because which service wrapped the key is not
// something the configuration can be trusted to still say: an operator who
// switches providers, or who restores an older backup, needs the stored value
// itself to be unambiguous.
const storedPrefix = "hpkms:"

// pack writes the stored form, which is the marker, the provider and its payload.
func pack(provider, payload string) string {
	return storedPrefix + provider + ":" + payload
}

// unpack reads the stored form, and reports whether it is one.
func unpack(stored string) (provider, payload string, ok bool) {
	rest, found := strings.CutPrefix(stored, storedPrefix)
	if !found {
		return "", "", false
	}
	provider, payload, found = strings.Cut(rest, ":")
	if !found || provider == "" {
		return "", "", false
	}
	return provider, payload, true
}

// ProviderOf names the provider that wrapped a stored key.
//
// A value with no marker was wrapped with the app secret, by the only scheme
// there was before this package, so it is local's. That is what keeps an
// installation that upgrades into this code starting without being touched.
func ProviderOf(stored string) string {
	if provider, _, ok := unpack(stored); ok {
		return provider
	}
	return ProviderLocal
}

// payloadFor reads the payload of a stored value this provider wrote, and
// refuses one that another provider wrote: unwrapping it would either fail
// obscurely or, worse, succeed on something that is not the key.
func payloadFor(provider, stored string) (string, error) {
	name, payload, ok := unpack(stored)
	if !ok {
		return "", fmt.Errorf("%w: expected a %s key, got one with no marker", ErrWrapped, provider)
	}
	if name != provider {
		return "", fmt.Errorf("%w: expected a %s key, got a %s one", ErrWrapped, provider, name)
	}
	return payload, nil
}
