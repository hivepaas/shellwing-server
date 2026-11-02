package licensesig

import (
	"crypto"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// maxSignatures must match releasesig's bound in HivePaaS, which refuses an envelope with more.
const maxSignatures = 16

// ErrKeyInvalid is a key whose envelope does not open: malformed, or without a valid signature in every required
// algorithm by the keys it was opened with.
var ErrKeyInvalid = errors.New("license key does not verify")

// envelope is a license key: the payload, base64, exactly the bytes signed, and their SHA-256. It is
// releasesig.Envelope in HivePaaS, field for field and in the same order, because the token is this struct's JSON.
type envelope struct {
	Payload    string  `json:"payload"`
	SHA256     string  `json:"sha256"`
	Signatures []entry `json:"signatures"`
}

type entry struct {
	KeyID     string `json:"keyId"`
	Algorithm string `json:"alg"`
	Sig       string `json:"sig"`
}

type publicKey struct {
	id  string
	alg string
	key crypto.PublicKey
}

// openKey applies HivePaaS's rule (releasesig.Purpose.Open under the signing context) to a key's text and returns the
// payload bytes: for every required algorithm, a valid signature by one of pubs. A signature by a key not in pubs is
// passed over; one by a key in pubs that does not verify, or claims another algorithm than the key has, refuses the
// key.
//
// The portal opens what it just signed, as a self-check, and a key it has no record of (Verifier), so this is the
// same logic as HivePaaS's and no more.
func openKey(context string, pubs []*publicKey, text string) ([]byte, error) {
	env, data, err := parseEnvelope(text)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]*publicKey, len(pubs))
	for _, pub := range pubs {
		byID[pub.id] = pub
	}
	satisfied := map[string]bool{}
	for _, sigEntry := range env.Signatures {
		pub, found := byID[sigEntry.KeyID]
		if !found {
			continue
		}
		if sigEntry.Algorithm != pub.alg {
			return nil, keyInvalid("key %q is %s, signature claims %q", sigEntry.KeyID, pub.alg, sigEntry.Algorithm)
		}
		sig, err := base64.StdEncoding.DecodeString(sigEntry.Sig)
		if err != nil || !verify(context, pub, data, sig) {
			return nil, keyInvalid("signature by key %q does not verify", sigEntry.KeyID)
		}
		satisfied[pub.alg] = true
	}
	for _, alg := range requiredAlgorithms {
		if !satisfied[alg] {
			return nil, keyInvalid("no valid %s signature by the given keys", alg)
		}
	}
	return data, nil
}

// parseEnvelope reads a key's text, in any form HivePaaS takes, to its envelope and the payload bytes it signs, and
// checks what holds without any key: the envelope's shape, and its sha256 against the payload. No signature is
// verified.
func parseEnvelope(text string) (*envelope, []byte, error) {
	raw, err := decodeKeyText(text)
	if err != nil {
		return nil, nil, err
	}
	var env envelope
	if err = json.Unmarshal(raw, &env); err != nil {
		return nil, nil, keyInvalid("malformed envelope")
	}
	if len(env.Signatures) > maxSignatures {
		return nil, nil, keyInvalid("too many signatures")
	}
	data, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil || len(data) == 0 {
		return nil, nil, keyInvalid("malformed payload")
	}
	sum := sha256.Sum256(data)
	if subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(env.SHA256)) != 1 {
		return nil, nil, keyInvalid("sha256 does not match the payload")
	}
	return &env, data, nil
}

func verify(context string, pub *publicKey, data, sig []byte) bool {
	switch key := pub.key.(type) {
	case ed25519.PublicKey:
		return ed25519.VerifyWithOptions(key, data, sig, &ed25519.Options{Context: context}) == nil
	case *mldsa.PublicKey:
		return mldsa.Verify(key, data, sig, &mldsa.Options{Context: context}) == nil
	default:
		return false
	}
}

// decodeKeyText is license.DecodeKeyText in HivePaaS: once whitespace is removed, the text is the envelope's JSON as
// base64 (with or without padding), which is the form keys are issued in, or the envelope's JSON as it is.
func decodeKeyText(text string) ([]byte, error) {
	text = NormalizeKeyText(text)
	if text == "" {
		return nil, keyInvalid("no key")
	}
	if len(text) > MaxKeyTextLen {
		return nil, keyInvalid("longer than %d characters", MaxKeyTextLen)
	}
	if strings.HasPrefix(text, "{") {
		return []byte(text), nil
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(text, "="))
	if err != nil || len(raw) == 0 {
		return nil, keyInvalid("neither base64 nor JSON")
	}
	return raw, nil
}

func keyInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrKeyInvalid, fmt.Sprintf(format, args...))
}
