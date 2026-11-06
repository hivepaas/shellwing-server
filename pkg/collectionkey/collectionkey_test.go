package collectionkey

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// counting is n bytes counting up from from.
func counting(from byte, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = from + byte(i)
	}
	return b
}

// vector is the wrap the app's format document lists: the identity vector's X25519 key as the recipient (secrets 1,
// 2, 3…), and parts counting up from 65, 97 and 169.
func vector() *Wrap {
	recipient, _ := base64.StdEncoding.DecodeString("B6N8vBQgk8i3VdwbEOhstCY3StFqqFPtC9/AsrhtHHw=")
	return &Wrap{
		CollectionID:    "01JC0LLECT10N0000000000000",
		Version:         1,
		RecipientX25519: recipient,
		EphemeralPublic: counting(65, KeyLen),
		Sealed:          counting(97, SealedLen),
		Commitment:      counting(169, KeyLen),
	}
}

// The identity vector's Ed25519 key (secrets 33, 34, 35…) signs the vector, the same every time.
func TestTheDocumentedVector(t *testing.T) {
	wrapper := ed25519.NewKeyFromSeed(counting(33, ed25519.SeedSize))
	w := vector()
	signature := ed25519.Sign(wrapper, w.Signed())
	assert.Equal(t, len(signedLabel)+2+26+4+32+32+72+32, len(w.Signed()))
	assert.Equal(t, "hV8F1bdORRJKFUVYFOyYGOBIeiybJv+zC3lm3p8xFzyZGcJ7Hz889YqDOoBYHSwhqRqO8J8JH4ExwMrNxSjkAw==",
		base64.StdEncoding.EncodeToString(signature))
	assert.True(t, w.Verify(wrapper.Public().(ed25519.PublicKey), signature))
}

func TestWhatIsSignedCannotChange(t *testing.T) {
	wrapper := ed25519.NewKeyFromSeed(counting(33, ed25519.SeedSize))
	public := wrapper.Public().(ed25519.PublicKey)
	signature := ed25519.Sign(wrapper, vector().Signed())

	for name, change := range map[string]func(w *Wrap){
		"collection": func(w *Wrap) { w.CollectionID = "01JC0LLECT10N0000000000001" },
		"version":    func(w *Wrap) { w.Version = 2 },
		"recipient":  func(w *Wrap) { w.RecipientX25519[0] ^= 1 },
		"ephemeral":  func(w *Wrap) { w.EphemeralPublic[0] ^= 1 },
		"sealed":     func(w *Wrap) { w.Sealed[71] ^= 1 },
		"commitment": func(w *Wrap) { w.Commitment[31] ^= 1 },
	} {
		w := vector()
		change(w)
		assert.False(t, w.Verify(public, signature), name)
	}

	short := vector()
	short.Sealed = short.Sealed[:SealedLen-1]
	assert.False(t, short.Verify(public, ed25519.Sign(wrapper, short.Signed())), "a sealed key of another size")
	unversioned := vector()
	unversioned.Version = 0
	assert.False(t, unversioned.Verify(public, ed25519.Sign(wrapper, unversioned.Signed())), "version 0")
}
