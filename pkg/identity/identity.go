// Package identity checks a person's identity as the app publishes it: an X25519 public key, which collection keys are
// wrapped for, an Ed25519 one, which signs what they wrap, and the X25519 key signed by the Ed25519 one. The server
// holds no secret of it. The app's side is shellwing's crates/vault/src/identity.rs; both test the same vector.
package identity

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	// KeyLen is the length of either public key.
	KeyLen = 32
	// SignatureLen is the length of the signature.
	SignatureLen = ed25519.SignatureSize

	signedLabel      = "shellwing identity signature v1"
	fingerprintLabel = "shellwing identity v1"
	fingerprintBytes = 30
	groupBytes       = 5
	groupModulus     = 100000
)

// Verify says whether the Ed25519 key signed the X25519 one: that one person holds both.
func Verify(x25519, ed25519Key, signature []byte) bool {
	if len(x25519) != KeyLen || len(ed25519Key) != ed25519.PublicKeySize || len(signature) != SignatureLen {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(ed25519Key), signed(x25519, ed25519Key), signature)
}

func signed(x25519, ed25519Key []byte) []byte {
	return append(append([]byte(signedLabel), x25519...), ed25519Key...)
}

// Fingerprint is what people compare to know they have each other's keys: six groups of five digits, from the first 30
// bytes of a hash of both public keys, each group the big-endian number of five bytes modulo 100000.
func Fingerprint(x25519, ed25519Key []byte) string {
	hash := sha256.Sum256(append(append([]byte(fingerprintLabel), x25519...), ed25519Key...))
	groups := make([]string, 0, fingerprintBytes/groupBytes)
	for i := 0; i < fingerprintBytes; i += groupBytes {
		var n [8]byte
		copy(n[8-groupBytes:], hash[i:i+groupBytes])
		groups = append(groups, fmt.Sprintf("%05d", binary.BigEndian.Uint64(n[:])%groupModulus))
	}
	return strings.Join(groups, " ")
}
