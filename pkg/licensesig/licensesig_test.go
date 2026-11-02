package licensesig

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The keys in testdata were signed by the portal's own signer, with its development keys and this product's context.
func testKey(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name + ".txt")
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}

func devVerifier(t *testing.T) *Verifier {
	t.Helper()
	verifier, err := Embedded(true)
	require.NoError(t, err)
	return verifier
}

func TestVerifiesAKeyThePortalSigned(t *testing.T) {
	payload, err := devVerifier(t).Verify(testKey(t, "team-10"))
	require.NoError(t, err)
	assert.Equal(t, "team-10", payload.Plan)
	assert.Equal(t, 10, payload.Limits[LimitUsers])
	assert.Equal(t, "Acme", payload.Customer.Name)
	assert.True(t, payload.ForInstallation("q9xk2m4j7n1p0r5s8t3v6w2yz"))
	assert.False(t, payload.ForInstallation("another"))
	assert.Equal(t, time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC), payload.GraceEndsAt())

	trial, err := devVerifier(t).Verify(testKey(t, "trial"))
	require.NoError(t, err)
	assert.True(t, trial.Trial)
}

func TestAKeyWrappedByAnEmailClientIsTheSameKey(t *testing.T) {
	key := testKey(t, "team-10")
	wrapped := key[:40] + "\n  " + key[40:80] + "\r\n" + key[80:] + "\n"
	_, err := devVerifier(t).Verify(wrapped)
	require.NoError(t, err)
}

func TestRefuses(t *testing.T) {
	verifier := devVerifier(t)
	key := testKey(t, "team-10")

	_, err := verifier.Verify(testKey(t, "hivepaas-key"))
	require.ErrorIs(t, err, ErrKeyInvalid, "another product's key, with the same signing keys")

	for name, text := range map[string]string{
		"nothing":  "",
		"not one":  "not a key",
		"cut":      key[:len(key)/2],
		"tampered": key[:100] + flip(key[100]) + key[101:],
	} {
		_, err = verifier.Verify(text)
		assert.ErrorIs(t, err, ErrKeyInvalid, name)
	}
}

func flip(c byte) string {
	if c == 'A' {
		return "B"
	}
	return "A"
}

// The development pair's private half is public: only a server in development may trust it.
func TestADevelopmentKeyIsRefusedOutsideDevelopment(t *testing.T) {
	verifier, err := Embedded(false)
	require.NoError(t, err)
	_, err = verifier.Verify(testKey(t, "team-10"))
	require.ErrorIs(t, err, ErrKeyInvalid)
}

// A production server trusts the pair the portal signs every product's keys with, and no other.
func TestAProductionServerTrustsTheSharedPair(t *testing.T) {
	verifier, err := Embedded(false)
	require.NoError(t, err)
	algs := map[string]string{}
	for _, key := range verifier.keys {
		algs[key.id] = key.alg
	}
	assert.Equal(t, map[string]string{"lic-2026-ed": AlgEd25519, "lic-2026-ml": AlgMLDSA65}, algs)
}

func TestValidate(t *testing.T) {
	good := func() *Payload {
		return &Payload{Version: 1, LicenseID: "l", Plan: "team-10", Installations: []string{"i"},
			IssuedAt: time.Unix(1, 0), ExpiresAt: time.Unix(2, 0), GraceDays: 30,
			Limits: map[string]int{LimitUsers: 10, "something-new": 1}}
	}
	require.NoError(t, validate(good()), "a limit this binary does not know is passed over")
	for name, change := range map[string]func(*Payload){
		"version":    func(p *Payload) { p.Version = 2 },
		"id":         func(p *Payload) { p.LicenseID = "" },
		"plan":       func(p *Payload) { p.Plan = "" },
		"no servers": func(p *Payload) { p.Installations = nil },
		"empty one":  func(p *Payload) { p.Installations = []string{""} },
		"dates":      func(p *Payload) { p.ExpiresAt = p.IssuedAt },
		"grace":      func(p *Payload) { p.GraceDays = MaxGraceDays + 1 },
		"negative":   func(p *Payload) { p.GraceDays = -1 },
		"limit":      func(p *Payload) { p.Limits[LimitUsers] = -1 },
	} {
		p := good()
		change(p)
		assert.ErrorIs(t, validate(p), ErrInvalidPayload, name)
	}
}
