// Package licensesig verifies the license keys the portal signs for a shellwing server. It holds no signing key and
// signs nothing: a server verifies offline, with the public keys in its binary.
//
// A license key is an envelope: the payload's JSON, base64, with an ed25519 and an ML-DSA-65 signature over it under
// the context "shellwing-server-license-v1", both required. As text it is the envelope's compact JSON as base64. The
// format is the portal's, whose bucket of rules this keeps to, and the test keys in testdata are ones it signed.
//
// It is a fair-use gate, not DRM: the code is open, so anybody may take the limit out for their own use.
package licensesig

import (
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"
)

const (
	// Context is the signing context of both algorithms: it keeps a key of another product from ever verifying here.
	Context = "shellwing-server-license-v1"

	AlgEd25519 = "ed25519"
	AlgMLDSA65 = "ml-dsa-65"

	// PayloadVersion is the version of a key this binary reads: a key of another version is refused, not half-read.
	PayloadVersion = 1
	// MaxGraceDays bounds a key's grace.
	MaxGraceDays = 365
	// LimitUsers is the name a key gives its limit of users.
	LimitUsers = "users"

	pubKeySuffix = ".pub.pem"
)

var requiredAlgorithms = []string{AlgEd25519, AlgMLDSA65}

var (
	// ErrInvalidPayload is a key that verifies but says something a key may not.
	ErrInvalidPayload = errors.New("license payload is invalid")
	// ErrPublicKeys is a set of public keys that cannot be read.
	ErrPublicKeys = errors.New("license public keys are invalid")
)

//go:embed licensekeys
var embedded embed.FS

// Payload is what a license key says. It holds the customer's email, so it is never logged and never put in an error.
type Payload struct {
	Version       int            `json:"version"`
	LicenseID     string         `json:"licenseId"`
	Customer      Customer       `json:"customer"`
	Plan          string         `json:"plan"`
	Limits        map[string]int `json:"limits,omitempty"`
	Installations []string       `json:"installations"`
	IssuedAt      time.Time      `json:"issuedAt"`
	ExpiresAt     time.Time      `json:"expiresAt"`
	GraceDays     int            `json:"graceDays"`
	Trial         bool           `json:"trial,omitempty"`
}

// Customer is who a key was issued to.
type Customer struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// GraceEndsAt is when the key stops holding.
func (p *Payload) GraceEndsAt() time.Time {
	return p.ExpiresAt.Add(time.Duration(p.GraceDays) * 24 * time.Hour)
}

// ForInstallation says whether the key names this server.
func (p *Payload) ForInstallation(id string) bool {
	return slices.Contains(p.Installations, id)
}

// A plan name or a limit this binary does not know is passed over rather than refused: tiers change at the portal,
// and a server already running must go on taking what it sells.
func validate(p *Payload) error {
	switch {
	case p.Version != PayloadVersion:
		return invalid("version must be %d", PayloadVersion)
	case p.LicenseID == "":
		return invalid("licenseId is required")
	case p.Plan == "":
		return invalid("plan is required")
	case len(p.Installations) == 0 || slices.Contains(p.Installations, ""):
		return invalid("an installation is required")
	case p.IssuedAt.IsZero() || p.ExpiresAt.IsZero() || !p.ExpiresAt.After(p.IssuedAt):
		return invalid("expiresAt must be after issuedAt")
	case p.GraceDays < 0 || p.GraceDays > MaxGraceDays:
		return invalid("graceDays must be between 0 and %d", MaxGraceDays)
	}
	for name, value := range p.Limits {
		if value < 0 {
			return invalid("limit %q must not be negative", name)
		}
	}
	return nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidPayload, fmt.Sprintf(format, args...))
}

// Verifier verifies keys with one set of public keys.
type Verifier struct {
	keys []*publicKey
}

// NewVerifier reads public keys, PKIX PEM by key ID. Each must be an ed25519 or ML-DSA-65 key.
func NewVerifier(pems map[string][]byte) (*Verifier, error) {
	keys := make([]*publicKey, 0, len(pems))
	for id, data := range pems {
		block, _ := pem.Decode(data)
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("%w: %q is no PUBLIC KEY PEM block", ErrPublicKeys, id)
		}
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrPublicKeys, id, err)
		}
		switch key := parsed.(type) {
		case ed25519.PublicKey:
			keys = append(keys, &publicKey{id: id, alg: AlgEd25519, key: key})
		case *mldsa.PublicKey:
			if key.Parameters() != mldsa.MLDSA65() {
				return nil, fmt.Errorf("%w: %q: ML-DSA parameter set %s", ErrPublicKeys, id, key.Parameters())
			}
			keys = append(keys, &publicKey{id: id, alg: AlgMLDSA65, key: key})
		default:
			return nil, fmt.Errorf("%w: %q: unsupported key type %T", ErrPublicKeys, id, parsed)
		}
	}
	return &Verifier{keys: keys}, nil
}

// Embedded is a verifier of the public keys in the binary: the production ones, and the development ones if allowed.
// The development pair's private half is public, so a server only trusts it while in development.
func Embedded(allowDevelopment bool) (*Verifier, error) {
	pems, err := embeddedKeys("licensekeys/prod")
	if err != nil {
		return nil, err
	}
	if allowDevelopment {
		dev, err := embeddedKeys("licensekeys/dev")
		if err != nil {
			return nil, err
		}
		for id, data := range dev {
			pems[id] = data
		}
	}
	return NewVerifier(pems)
}

func embeddedKeys(dir string) (map[string][]byte, error) {
	entries, err := fs.ReadDir(embedded, dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPublicKeys, err)
	}
	pems := map[string][]byte{}
	for _, entry := range entries {
		id, isKey := strings.CutSuffix(entry.Name(), pubKeySuffix)
		if !isKey {
			continue
		}
		data, err := embedded.ReadFile(path.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPublicKeys, err)
		}
		pems[id] = data
	}
	return pems, nil
}

// Verify checks a key's signatures, a valid one in every required algorithm, and reads what it says. Anything else is
// ErrKeyInvalid or ErrInvalidPayload; neither quotes the payload.
func (v *Verifier) Verify(text string) (*Payload, error) {
	data, err := openKey(Context, v.keys, text)
	if err != nil {
		return nil, err
	}
	var payload Payload
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, keyInvalid("malformed payload")
	}
	if err = validate(&payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
