package sessionuc

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A code goes nowhere but an app's own loopback listener, or a web page could have the browser hand it over.
func TestValidDeviceRedirect(t *testing.T) {
	for _, good := range []string{"http://127.0.0.1:53124/callback", "http://[::1]:53124/callback"} {
		assert.True(t, validDeviceRedirect(good), good)
	}
	for _, bad := range []string{
		"",
		"https://evil.example.com/callback",
		"http://evil.example.com:53124/callback",
		"http://localhost:53124/callback",
		"http://127.0.0.1.evil.example.com:53124/callback",
		"http://127.0.0.1:53124@evil.example.com/callback",
		"http://user@127.0.0.1:53124/callback",
		"http://127.0.0.1/callback",
		"http://127.0.0.1:0/callback",
		"http://127.0.0.1:53124/other",
		"http://127.0.0.1:53124/callback?x=1",
		"http://127.0.0.1:53124/callback#x",
		"https://127.0.0.1:53124/callback",
		"http://10.0.0.1:53124/callback",
		"javascript:alert(1)",
		"http://127.0.0.1:53124//evil.example.com/callback",
	} {
		assert.False(t, validDeviceRedirect(bad), bad)
	}
}

func TestPkceMatches(t *testing.T) {
	// RFC 7636, appendix B.
	assert.True(t, pkceMatches("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"))

	verifier := "a-verifier-long-enough-to-be-one-of-those-rfc7636-asks-for"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	assert.True(t, pkceMatches(verifier, challenge))
	assert.False(t, pkceMatches(verifier+"x", challenge))
	assert.False(t, pkceMatches(verifier, ""))
}
