// Package collectionkey checks a collection's key wrapped for a person, as the app makes it
// (docs/design/collections.md in the app's repository). The server never opens one: it checks who signed it, and for
// whom.
package collectionkey

import (
	"crypto/ed25519"
	"encoding/binary"
	"math"
)

const (
	// KeyLen is an X25519 public key's size, and the commitment's.
	KeyLen = 32
	// SealedLen is the sealed key: a 24-byte XChaCha20-Poly1305 nonce, the 32-byte key, and the 16-byte tag.
	SealedLen = 24 + 32 + 16
	// SignatureLen is an Ed25519 signature's size.
	SignatureLen = ed25519.SignatureSize
)

// signedLabel starts what the wrapper signs.
const signedLabel = "shellwing collection key wrap v1"

// Wrap is a collection's key wrapped for one person.
type Wrap struct {
	CollectionID string
	Version      int
	// RecipientX25519 is the X25519 key of the person it is wrapped for.
	RecipientX25519 []byte
	// EphemeralPublic is the public half of the key pair made for this wrap.
	EphemeralPublic []byte
	// Sealed is the collection key, sealed: nonce, then ciphertext and tag.
	Sealed []byte
	// Commitment is SHA-256("shellwing collection key commitment v1" ‖ key).
	Commitment []byte
}

// Signed is what the wrapper's Ed25519 key signs: the label, the collection's id as a 2-byte big-endian length then
// its bytes, the version as 4 bytes big-endian, the recipient's X25519 key, the ephemeral public key, the sealed key
// and the commitment.
func (w *Wrap) Signed() []byte {
	id := w.CollectionID
	out := make([]byte, 0, len(signedLabel)+2+len(id)+4+3*KeyLen+SealedLen)
	out = append(out, signedLabel...)
	out = binary.BigEndian.AppendUint16(out, uint16(min(len(id), math.MaxUint16))) //nolint:gosec // bounded
	out = append(out, id...)
	out = binary.BigEndian.AppendUint32(out, uint32(max(0, min(w.Version, math.MaxInt32)))) //nolint:gosec // bounded
	out = append(out, w.RecipientX25519...)
	out = append(out, w.EphemeralPublic...)
	out = append(out, w.Sealed...)
	out = append(out, w.Commitment...)
	return out
}

// WellFormed says whether each part has its size.
func (w *Wrap) WellFormed() bool {
	return w.CollectionID != "" && len(w.CollectionID) <= math.MaxUint16 && w.Version > 0 &&
		len(w.RecipientX25519) == KeyLen && len(w.EphemeralPublic) == KeyLen &&
		len(w.Sealed) == SealedLen && len(w.Commitment) == KeyLen
}

// Verify says whether the wrap is well formed and the Ed25519 key signed it.
func (w *Wrap) Verify(wrapperEd25519, signature []byte) bool {
	if !w.WellFormed() || len(wrapperEd25519) != ed25519.PublicKeySize || len(signature) != SignatureLen {
		return false
	}
	return ed25519.Verify(wrapperEd25519, w.Signed(), signature)
}
