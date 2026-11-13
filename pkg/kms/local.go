package kms

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hivepaas/shellwing-server/pkg/cryptoutil"
)

// wrapSaltLen is the salt length used when wrapping with the app secret. It is
// the length the scheme has always used, and changing it would make keys already
// stored unreadable.
const wrapSaltLen = 10

// local wraps the data encryption key with the app secret from the
// configuration.
//
// Its stored form carries no marker, which is deliberate: it is the form every
// installation already has in its database, from before a KMS was an option, and
// an upgrade must keep reading it without being asked to do anything. See
// ProviderOf.
type local struct {
	appSecret string
}

// NewLocal wraps with the app secret. It is slow on purpose - the secret is
// chosen by whoever deploys the server and may be weak, so the key is derived
// from it with Argon2id rather than used directly. That cost is paid once per
// start, not once per value.
func NewLocal(appSecret string) (Wrapper, error) {
	if appSecret == "" {
		return nil, fmt.Errorf("%w: %s needs the app secret", ErrNotConfigured, ProviderLocal)
	}
	return &local{appSecret: appSecret}, nil
}

func (w *local) Name() string { return ProviderLocal }

func (w *local) Wrap(_ context.Context, key []byte) (string, error) {
	wrapped, err := cryptoutil.EncryptBase64(
		base64.StdEncoding.EncodeToString(key), wrapSaltLen, w.appSecret)
	if err != nil {
		return "", fmt.Errorf("failed to wrap the data encryption key with the app secret: %w", err)
	}
	return wrapped, nil
}

func (w *local) Unwrap(_ context.Context, stored string) ([]byte, error) {
	// A marked value was wrapped by a service, and the app secret will not open
	// it. Say so, rather than letting the derivation fail with nothing to go on.
	if provider, _, ok := unpack(stored); ok {
		return nil, fmt.Errorf("%w: the stored key was wrapped by %s, not by the app secret",
			ErrWrapped, provider)
	}
	encoded, err := cryptoutil.DecryptBase64(stored, w.appSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap the data encryption key with the app secret: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWrapped, err)
	}
	return key, nil
}
